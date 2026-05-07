package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/chivta/ruscan/internal/shared/domain"
)

type JobRepo struct {
	redis *redis.Client
}

func NewJobRepo(redis *redis.Client) *JobRepo {
	return &JobRepo{redis: redis}
}

func jobKey(jobID string) string { return "jobs:" + jobID }

func (r *JobRepo) CreateJob(ctx context.Context, jobID, userID string) error {
	pipe := r.redis.Pipeline()
	pipe.HSet(ctx, jobKey(jobID),
		"status", string(domain.JobStatusPending),
		"created_at", time.Now().UTC().Format(time.RFC3339),
		"user_id", userID,
	)
	pipe.Expire(ctx, jobKey(jobID), domain.JobTTL)
	_, err := pipe.Exec(ctx)
	return err
}

func (r *JobRepo) GetJob(ctx context.Context, jobID, userID string) (*domain.Job, error) {
	fields, err := r.redis.HGetAll(ctx, jobKey(jobID)).Result()
	if err != nil {
		return nil, err
	}
	if len(fields) == 0 || fields["user_id"] != userID {
		return nil, domain.ErrNotFound
	}

	job := &domain.Job{
		JobID:  jobID,
		Status: domain.JobStatus(fields["status"]),
	}

	if t, err := time.Parse(time.RFC3339, fields["created_at"]); err == nil {
		job.CreatedAt = t
	}

	if data := fields["data"]; data != "" {
		job.Result = json.RawMessage(data)
	}

	return job, nil
}

func (r *JobRepo) PostJobResult(ctx context.Context, jobID string, status domain.JobStatus, result any) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}
	pipe := r.redis.Pipeline()
	pipe.HSet(ctx, jobKey(jobID),
		"status", string(status),
		"data", string(data),
	)
	pipe.Expire(ctx, jobKey(jobID), domain.JobTTL)
	_, err = pipe.Exec(ctx)
	return err
}

func (r *JobRepo) MarkJobFailed(ctx context.Context, jobID string, status domain.JobStatus, failReason domain.AppError) error {
	pipe := r.redis.Pipeline()
	pipe.HSet(ctx, jobKey(jobID),
		"status", string(status),
		"data", failReason.Code,
	)
	pipe.Expire(ctx, jobKey(jobID), domain.JobTTL)
	_, err := pipe.Exec(ctx)
	return err
}
