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
  scan: "Scan",
  placeholder: "Paste Spotify URL or ID...",
  scanning: "Scanning...",
  invalidInput: "Invalid URL or ID. Please check and try again.",
  anonQuotaExceeded: "You've used your trial scans. Sign in to keep scanning, it's free!",
  notFound: "Not found. Check the URL or ID and try again.",
  notFoundHints: [
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
  contentClear: "No russian content found 👏👏👏",
  noScanResults: "No scan results yet",
  noScanResultsHint: "Enter a Spotify URL or ID to get started",
  resourceType: {
    playlist: "Playlist",
    track: "Track",
    album: "Album",
    artist: "Artist",
  },

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
    scan: "Сканувати",
    placeholder: "Вставте посилання Spotify або ID...",
    scanning: "Сканування...",
    invalidInput: "Неправильне посилання або ID. Перевірте та спробуйте ще раз.",
    anonQuotaExceeded: "Ви використали пробні сканування. Увійдіть, щоб продовжити — це безкоштовно!",
    notFound: "Не знайдено. Перевірте посилання або ID та спробуйте ще раз.",
    notFoundHints: [
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
    contentClear: "Російського контенту не знайдено 👏👏👏",
    noScanResults: "Результатів сканування ще немає",
    noScanResultsHint: "Введіть посилання Spotify або ID, щоб розпочати",
    resourceType: {
      playlist: "Плейлист",
      track: "Трек",
      album: "Альбом",
      artist: "Артист",
    },

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
