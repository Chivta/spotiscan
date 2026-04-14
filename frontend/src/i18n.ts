export type Lang = "en" | "uk";

const trackCountUK = (n: number): string => {
  const mod10 = n % 10;
  const mod100 = n % 100;
  if (mod10 === 1 && mod100 !== 11) return `${n} трек`;
  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 10 || mod100 >= 20)) return `${n} треки`;
  return `${n} треків`;
};

const en = {
  // Landing
  signIn: "Sign in",
  createAccount: "Create account",
  landingHeadline: "Scan your Spotify music to find and remove tracks by Russian artists.",

  // Header
  logOut: "Log out",

  // Scanner
  scanner: "Scanner",
  scan: "Scan",
  placeholder: "Spotify URL, playlist ID, or artist name...",
  scanning: "Scanning...",
  invalidInput: "Enter a Spotify URL, URI, playlist ID, or artist name.",
  artistNotInBase: "This artist is not in our database.",
  anonQuotaExceeded: "You've used your trial scans. Sign in to keep scanning, it's free!",
  spotifyNotFound: "Not found on Spotify. Check the URL or name and try again.",
  spotifyNotFoundHints: [
    "Make sure the playlist is not set to private.",
    "Spotify restricts API access to algorithm-generated playlists — this includes Discover Weekly, Release Radar, and genre playlists owned by Spotify. You can copy the playlist contents into your own playlist and scan that instead.",
    "Liked Songs cannot be scanned directly either — copy them into a regular playlist first.",
  ] as string[],
  badRequest: "Invalid request. Please check your input.",
  databaseError: "A server error occurred. Please try again later.",
  spotifyApiError: "Failed to communicate with Spotify. Please try again later.",
  internalError: "An unexpected error occurred. Please try again later.",
  tooManyRequests: "Too many requests. Please wait a few minutes and try again.",
  unauthorized: "You need to be signed in to do that.",
  forbidden: "You don't have permission to do that.",
  notFound: "Not found.",
  somethingWentWrong: "Something went wrong",
  showingCachedResults: "Showing results from a previous scan.",
  rescan: "Rescan?",
  tracksTab: (n: number) => `Tracks (${n})`,
  artistsTab: (n: number) => `Artists (${n})`,
  searchTracks: "Search tracks...",
  tracksWithRussianArtists: (n: number) => `Tracks with Russian Artists (${n})`,
  russianArtistsFound: (n: number) => `Russian Artists found (${n})`,
  noRussianTracks: "No Russian tracks found. This content is clean!",
  noTracksMatch: "No tracks match your search",
  searchArtists: "Search artists...",
  noArtistsFound: "No artists found",
  noArtistsMatch: "No artists match your search",
  source: "Source:",
  confirmed: "confirmed",
  trackCount: (n: number) => `${n} track${n !== 1 ? "s" : ""}`,
  dataProvidedBySpotify: "Data provided by Spotify",
  scanResult: "Scan result",
  contentClear: "No Russian content found 👏👏👏",
  noScanResults: "No scan results yet",
  noScanResultsHint: "Enter a Spotify URL or ID to get started",
  resourceType: {
    playlist: "Playlist",
    track: "Track",
    album: "Album",
    artist: "Artist",
  },

  // Dashboard tabs
  tabScanner: "Scanner",
  tabSuggestions: "Suggest",

  // Artist Suggestions
  suggestionsTitle: "Suggest Artist",
  suggestionsHint: "Suggest an artist to be added to our Russian artists database.",
  suggestArtist: "Suggest an artist",
  artistNameLabel: "Artist name",
  artistNamePlaceholder: "Artist name on Spotify...",
  descriptionLabel: "Why is this artist Russian?",
  descriptionPlaceholder: "Provide evidence or reasoning (max 1000 characters)...",
  submitSuggestion: "Submit suggestion",
  submitting: "Submitting...",
  noSuggestions: "You haven't submitted any suggestions yet.",
  approved: "Approved",
  pending: "Pending",
  editSuggestion: "Edit",
  deleteSuggestion: "Delete",
  saveChanges: "Save",
  cancelEdit: "Cancel",
  confirmDelete: "Delete this suggestion?",
  suggestionCreated: "Suggestion submitted successfully.",
  suggestionUpdated: "Suggestion updated.",
  suggestionDeleted: "Suggestion deleted.",
  yourSuggestions: "Your suggestions",
  artistExists: "This artist is already in our database.",
  suggestionApproved: "This suggestion has already been approved and cannot be deleted.",

  // AuthPage
  logIn: "Log In",
  signUp: "Sign Up",
  emailPlaceholder: "Email",
  passwordPlaceholder: "Password",
  incorrectCredentials: "Incorrect email or password.",
  emailExists: "An account with this email already exists.",
  checkEmailPassword: "Please check your email and password.",
  tryAgain: "Something went wrong. Please try again.",
  loggingIn: "Logging in…",
  signingUp: "Signing up…",
  alreadyHaveAccount: "Already have an account? ",
  dontHaveAccount: "Don't have an account? ",
  logInLink: "Log in",
  signUpLink: "Sign up",
};

export type T = typeof en;

export const translations = {
  en,
  uk: {
    // Landing
    signIn: "Увійти",
    createAccount: "Створити акаунт",
    landingHeadline: "Скануйте музику у Spotify, щоб знайти та прибрати треки від російських артистів.",

    // Header
    logOut: "Вийти",

    // Scanner
    scanner: "Сканер",
    scan: "Сканувати",
    placeholder: "Посилання Spotify, ID плейлисту або ім'я артиста...",
    scanning: "Сканування...",
    invalidInput: "Введіть посилання Spotify, URI, ID плейлисту або ім'я артиста.",
    artistNotInBase: "Цього артиста немає в нашій базі.",
    anonQuotaExceeded: "Ви використали пробні сканування. Увійдіть, щоб продовжити — це безкоштовно!",
    spotifyNotFound: "Не знайдено у Spotify. Перевірте посилання або ID та спробуйте ще раз.",
    spotifyNotFoundHints: [
      "Переконайтеся, що плейлист не приватний.",
      "Spotify обмежує доступ через API до плейлистів, згенерованих їхніми алгоритмами — зокрема Discover Weekly, Release Radar та жанрових плейлистів, власником яких є Spotify. Ви можете скопіювати вміст такого плейлиста у свій власний і сканувати його.",
      "Збережені треки (Liked Songs) також не можна сканувати напряму — спочатку скопіюйте їх у звичайний плейлист.",
    ] as string[],
    badRequest: "Некоректний запит. Будь ласка, перевірте введені дані.",
    databaseError: "Помилка сервера. Будь ласка, спробуйте пізніше.",
    spotifyApiError: "Не вдалося зв'язатися зі Spotify. Будь ласка, спробуйте пізніше.",
    internalError: "Виникла непередбачена помилка. Будь ласка, спробуйте пізніше.",
    tooManyRequests: "Забагато запитів. Зачекайте декілька хвилин та спробуйте ще раз.",
    unauthorized: "Для цього потрібно увійти в акаунт.",
    forbidden: "У вас немає дозволу на цю дію.",
    notFound: "Нічого не знайдено.",
    somethingWentWrong: "Щось пішло не так",
    showingCachedResults: "Показано результати попереднього сканування.",
    rescan: "Сканувати знову?",
    tracksTab: (n: number) => `Треки (${n})`,
    artistsTab: (n: number) => `Артисти (${n})`,
    searchTracks: "Пошук треків...",
    tracksWithRussianArtists: (n: number) => `Треки з російськими артистами (${n})`,
    russianArtistsFound: (n: number) => `Російських артистів знайдено (${n})`,
    noRussianTracks: "Російських треків не знайдено. Контент чистий!",
    noTracksMatch: "Треків за вашим запитом не знайдено",
    searchArtists: "Пошук артистів...",
    noArtistsFound: "Артистів не знайдено",
    noArtistsMatch: "Артистів за вашим запитом не знайдено",
    source: "Джерело:",
    confirmed: "підтверджено",
    trackCount: trackCountUK,
    dataProvidedBySpotify: "Дані надано Spotify",
    scanResult: "Результат сканування",
    contentClear: "Російського контенту не знайдено 👏👏👏",
    noScanResults: "Результатів сканування ще немає",
    noScanResultsHint: "Введіть посилання Spotify або ID, щоб розпочати",
    resourceType: {
      playlist: "Плейлист",
      track: "Трек",
      album: "Альбом",
      artist: "Артист",
    },

    // Dashboard tabs
    tabScanner: "Сканер",
    tabSuggestions: "Запропонувати",

    // Artist Suggestions
    suggestionsTitle: "Запропонувати Артиста",
    suggestionsHint: "Запропонуйте артиста для додавання до нашої бази російських артистів.",
    suggestArtist: "Запропонувати артиста",
    artistNameLabel: "Ім'я артиста",
    artistNamePlaceholder: "Ім'я артиста у Spotify...",
    descriptionLabel: "Чому цей артист російський?",
    descriptionPlaceholder: "Надайте докази або обґрунтування (до 1000 символів)...",
    submitSuggestion: "Надіслати пропозицію",
    submitting: "Надсилання...",
    noSuggestions: "Ви ще не надсилали пропозицій.",
    approved: "Схвалено",
    pending: "На розгляді",
    editSuggestion: "Редагувати",
    deleteSuggestion: "Видалити",
    saveChanges: "Зберегти",
    cancelEdit: "Скасувати",
    confirmDelete: "Видалити цю пропозицію?",
    suggestionCreated: "Пропозицію успішно надіслано.",
    suggestionUpdated: "Пропозицію оновлено.",
    suggestionDeleted: "Пропозицію видалено.",
    yourSuggestions: "Ваші пропозиції",
    artistExists: "Цей артист вже є в нашій базі.",
    suggestionApproved: "Цю пропозицію вже схвалено — видалити її неможливо.",

    // AuthPage
    logIn: "Увійти",
    signUp: "Зареєструватися",
    emailPlaceholder: "Email",
    passwordPlaceholder: "Пароль",
    incorrectCredentials: "Неправильний email або пароль.",
    emailExists: "Акаунт з таким email вже існує.",
    checkEmailPassword: "Будь ласка, перевірте email та пароль.",
    tryAgain: "Щось пішло не так. Будь ласка, спробуйте ще раз.",
    loggingIn: "Вхід…",
    signingUp: "Реєстрація…",
    alreadyHaveAccount: "Вже маєте акаунт? ",
    dontHaveAccount: "Немає акаунту? ",
    logInLink: "Увійти",
    signUpLink: "Зареєструватися",
  },
} satisfies Record<Lang, T>;
