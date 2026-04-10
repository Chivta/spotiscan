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
  landingHeadline: "Scan your Spotify playlists to find and remove tracks by Russian artists.",

  // Header
  logOut: "Log out",

  // PlaylistScanner
  scanPlaylist: "Scan Playlist",
  playlistPlaceholder: "Paste playlist URL or ID...",
  scanning: "Scanning...",
  invalidPlaylistId: "Invalid playlist ID. Check the URL or ID and try again.",
  anonQuotaExceeded: "You've used your trial scans. Sign in to keep scanning, it's free!",
  playlistNotFound: "Playlist not found. Check the URL or ID and try again.",
  badRequest: "Invalid request. Please check your input.",
  databaseError: "A server error occurred. Please try again later.",
  spotifyApiError: "Failed to communicate with Spotify. Please try again later.",
  internalError: "An unexpected error occurred. Please try again later.",
  somethingWentWrong: "Something went wrong",
  showingCachedResults: "Showing results from a previous scan.",
  rescan: "Rescan?",
  tracksTab: (n: number) => `Tracks (${n})`,
  artistsTab: (n: number) => `Artists (${n})`,
  searchTracks: "Search tracks...",
  tracksWithRussianArtists: (n: number) => `Tracks with Russian Artists (${n})`,
  russianArtistsFound: (n: number) => `Russian Artists found in playlist (${n})`,
  noRussianTracks: "No Russian tracks found. This playlist is clean!",
  noTracksMatch: "No tracks match your search",
  searchArtists: "Search artists...",
  noArtistsFound: "No artists found",
  noArtistsMatch: "No artists match your search",
  source: "Source:",
  confirmed: "confirmed",
  trackCount: (n: number) => `${n} track${n !== 1 ? "s" : ""}`,
  dataProvidedBySpotify: "Data provided by Spotify",
  noScanResults: "No scan results yet",
  noScanResultsHint: "Enter a playlist ID or URL to get started",

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
    landingHeadline: "Скануйте плейлисти Spotify, щоб знайти та прибрати треки від російських артистів.",

    // Header
    logOut: "Вийти",

    // PlaylistScanner
    scanPlaylist: "Сканувати плейлист",
    playlistPlaceholder: "Вставте посилання або ID плейлиста...",
    scanning: "Сканування...",
    invalidPlaylistId: "Неправильний ID плейлиста. Перевірте посилання або ID та спробуйте ще раз.",
    anonQuotaExceeded: "Ви використали пробні сканування. Увійдіть, щоб продовжити — це безкоштовно!",
    playlistNotFound: "Плейлист не знайдено. Перевірте посилання або ID та спробуйте ще раз.",
    badRequest: "Некоректний запит. Будь ласка, перевірте введені дані.",
    databaseError: "Помилка сервера. Будь ласка, спробуйте пізніше.",
    spotifyApiError: "Не вдалося зв'язатися зі Spotify. Будь ласка, спробуйте пізніше.",
    internalError: "Виникла непередбачена помилка. Будь ласка, спробуйте пізніше.",
    somethingWentWrong: "Щось пішло не так",
    showingCachedResults: "Показано результати попереднього сканування.",
    rescan: "Сканувати знову?",
    tracksTab: (n: number) => `Треки (${n})`,
    artistsTab: (n: number) => `Артисти (${n})`,
    searchTracks: "Пошук треків...",
    tracksWithRussianArtists: (n: number) => `Треки з російськими артистами (${n})`,
    russianArtistsFound: (n: number) => `Російські артисти у плейлисті (${n})`,
    noRussianTracks: "Російських треків не знайдено. Плейлист чистий!",
    noTracksMatch: "Треків за вашим запитом не знайдено",
    searchArtists: "Пошук артистів...",
    noArtistsFound: "Артистів не знайдено",
    noArtistsMatch: "Артистів за вашим запитом не знайдено",
    source: "Джерело:",
    confirmed: "підтверджено",
    trackCount: trackCountUK,
    dataProvidedBySpotify: "Дані надано Spotify",
    noScanResults: "Результатів сканування ще немає",
    noScanResultsHint: "Введіть ID або посилання на плейлист, щоб розпочати",

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
