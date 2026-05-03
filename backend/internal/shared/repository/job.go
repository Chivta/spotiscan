package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/chivta/ruscan/internal/shared/appErrors"
	"github.com/chivta/ruscan/internal/shared/models"
)

type JobRepo struct {
	redis *redis.Client
}

func NewJobRepo(redis *redis.Client) *JobRepo {
	return &JobRepo{redis: redis}
}

func jobKey(jobID string) string { return "jobs:" + jobID }

func (r *JobRepo) CreateJob(ctx context.Context, jobID string) error {
	pipe := r.redis.Pipeline()
	pipe.HSet(ctx, jobKey(jobID),
		"status", string(models.JobStatusPending),
		"created_at", time.Now().UTC().Format(time.RFC3339),
	)
	pipe.Expire(ctx, jobKey(jobID), models.JobTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *JobRepo) GetJob(ctx context.Context, jobID string) (*models.Job, error) {
	fields, err := r.redis.HGetAll(ctx, jobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 {
		return nil, appErrors.ErrNotFound
	}

	job := &models.Job{
		JobID:  jobID,
		Status: models.JobStatus(fields["status"]),
		Error:  fields["error"],
	}

	if t, err := time.Parse(time.RFC3339, fields["created_at"]); err == nil {
		job.CreatedAt = t
	}

	if data := fields["data"]; data != "" {
		job.Result = json.RawMessage(data)
	}

	return job, nil
}

func (r *JobRepo) PostJobResult(ctx context.Context, jobID string, ruContent *models.RuContent) error {
	data, err := json.Marshal(ruContent)
	if err != nil {
		return err
	}
	pipe := r.redis.Pipeline()
	pipe.HSet(ctx, jobKey(jobID),
		"status", string(models.JobStatusDone),
		"data", string(data),
	)
	pipe.Expire(ctx, jobKey(jobID), models.JobTTL)
	_, err = pipe.Exec(ctx)
	return err
}
