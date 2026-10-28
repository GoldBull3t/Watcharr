export type WatchedStatus =
	"PLANNED" | "WATCHING" | "FINISHED" | "HOLD" | "DROPPED" | "CANCELLED";
/**
 * Types of media supported by Watcharr in an overarching way.
 */
export type SupportedMedia = "tv" | "movie" | "game";
export type ContentType = SupportedMedia | "tv_episode";
export type MediaType = ContentType | "person";

// Wasn't able to figure out how to import this type from its component file in other places, so its here for now.
export type Icon =
	| "check"
	| "clock"
	| "calendar"
	| "database"
	| "archive"
	| "download"
	| "upload"
	| "thumb-down"
	| "thumb-up"
	| "play"
	| "pause"
	| "jellyfin"
	| "emby"
	| "plex"
	| "trash"
	| "close"
	| "close-circle"
	| "filter"
	| "filter-circle"
	| "reel"
	| "compass"
	| "document"
	| "add"
	| "arrow"
	| "chevron"
	| "search"
	| "sort"
	| "eye-closed"
	| "people-nocircle"
	| "people"
	| "person"
	| "person-add"
	| "person-minus"
	| "pencil"
	| "eye"
	| "star"
	| "movary"
	| "ryot"
	| "trakt"
	| "myanimelist"
	| "todomovies"
	| "themoviedb"
	| "imdb"
	| "refresh"
	| "undo"
	| "gamepad"
	| "film"
	| "tv"
	| "pin"
	| "unpin"
	| "sparkles"
	| "tag"
	| "ticket"
	| "lock-closed"
	| "github"
	| "website"
	| "tmdb"
	| "igdb"
	| "yamtrack"
	| "letterboxd"
	| "download"
	| "settings"
	| "notifications"
	| "discord";

export type Theme = "light" | "dark" | "system";

export type WLDetailedViewOption =
	"statusRating" | "lastWatched" | "dateAdded" | "dateModified";

export enum UserType {
	Watcharr = 0,
	Jellyfin = 1,
	Plex = 2,
	Proxy = 3,
}

export interface AuthResponse {
	token: string;
}

interface dbModel {
	id: number;
	createdAt: string;
	updatedAt: string;
	deletedAt: string;
}

export interface PaginationResponse<T, U> {
	limit: number;
	page: number;
	totalPages: number;
	totalResults: number;
	results: T[] | null;
	meta?: U;
}

export interface Content {
	// id: number; // Not used
	tmdbId: number;
	title: string;
	poster_path: string;
	overview: string;
	type: ContentType;
	release_date: string;
	first_air_date: string;
}

export type ActivityType =
	| "ADDED_WATCHED"
	| "REMOVED_WATCHED"
	| "RATING_CHANGED"
	| "STATUS_CHANGED"
	| "THOUGHTS_CHANGED"
	| "THOUGHTS_REMOVED"
	| "IMPORTED_RATING"
	| "IMPORTED_ADDED_WATCHED"
	| "SEASON_ADDED"
	| "SEASON_REMOVED"
	| "SEASON_RATING_CHANGED"
	| "SEASON_STATUS_CHANGED"
	| "EPISODE_ADDED"
	| "EPISODE_REMOVED"
	| "EPISODE_RATING_CHANGED"
	| "EPISODE_STATUS_CHANGED";

export enum ActivityCreatedBy {
	Watcharr = 1,
	GenericImport,
	JellyfinImport,
	JellyfinWebhook,
	PlexImport,
}

export interface Activity extends dbModel {
	watchedId: number;
	type: ActivityType;
	data: string;
	customDate: string;
	countAsPlay: boolean;
	createdBy?: ActivityCreatedBy;
	reason?: string;
}

export interface WatchedSeason {
	id: number;
	watchedID: number;
	status: WatchedStatus;
	rating: number;
	seasonNumber: number;
}

export interface WatchedEpisode {
	id: number;
	watchedID: number;
	status: WatchedStatus;
	rating: number;
	seasonNumber: number;
	episodeNumber: number;
}

/**
 * Same as `WatchedDto` on server.
 */
export interface Watched {
	id: number;
	createdAt: string;
	updatedAt: string;

	rating?: number;
	status: WatchedStatus;
	thoughts: string;
	pinned: boolean;

	activity?: Activity[];
	watchedSeasons?: WatchedSeason[];
	watchedEpisodes?: WatchedEpisode[];
	tags?: Tag[];
	lastViewedSeason?: number;
	plays?: number;

	// 'Watching Season/Ep' Extra detail.
	watchingSeason?: string;

	media?: Media;
}

export interface WatchedAddRequest {
	tmdbId?: number;
	igdbId?: number;
	contentType: SupportedMedia;
	rating?: number;
	status?: WatchedStatus;
}

export interface WatchedUpdateRequest {
	rating?: number;
	status?: WatchedStatus;
	thoughts?: string;
	removeThoughts?: boolean;
	pinned?: boolean;
	letCountAsPlay?: boolean;
}

export interface WatchedUpdateResponse {
	newActivities: Activity[];
}

export type WatchedSort =
	"DATEADDED" | "LASTCHANGED" | "LASTFIN" | "RATING" | "ALPHA" | "DATERELEASED";

export type SortDirection = "asc" | "desc";

export interface WatchedGetPageRequest {
	// Sorting type.
	sort?: WatchedSort;
	// Sorting direction (asc or desc).
	sortDir?: SortDirection;
	// Filtering options.
	type?: SupportedMedia[];
	status?: WatchedStatus[];
	year?: number;
	minRating?: number;
	maxRating?: number;
	genre?: string;
	query?: string;
}

export interface ActivityUpdateRequest {
	customDate: string;
}

export interface WatchedSeasonSetResponse {
	watchedSeason: WatchedSeason;
	update: boolean;
	addedActivities?: Activity[];
	statusChangedHookResponse?: SeasonStatusChangedHookResponse;
	watchedEpisodes?: WatchedEpisode[];
}

export interface WatchedEpisodeSetResponse {
	watchedEpisode: WatchedEpisode;
	update: boolean;
	addedActivities?: Activity[];
	statusChangedHookResponse?: EpisodeStatusChangedHookResponse;
}

export interface EpisodeStatusChangedHookResponse {
	newShowStatus?: WatchedStatus;
	watchedSeason?: WatchedSeason;
	addedActivities?: Activity[];
	errors?: string[];
}

export interface SeasonStatusChangedHookResponse {
	newShowStatus?: WatchedStatus;
	addedActivities?: Activity[];
	errors?: string[];
	watchedEpisodes?: WatchedEpisode[];
}

export interface RecentActivityItem {
	id: number;
	type: ActivityType;
	data: string;
	date: string;
	mediaTitle: string;
	mediaType: string;
	mediaId: number;
	posterPath?: string;
}

export interface ActivityPeriod {
	period: string;
	count: number;
	movies: number;
	episodes: number;
}

export interface RatingStats {
	count: number;
	average: number;
	highest: number;
	lowest: number;
	distribution: Record<string, number>;
}

export interface GenreStat {
	name: string;
	count: number;
}

export interface Profile {
	joined: Date | string;
	showsWatched: number;
	moviesWatched: number;
	moviesWatchedRuntime: number;
	showsWatchedRuntime: number;
	showsWatching: number;
	episodesWatched: number;
	gamesPlayed: number;
	totalRuntime: number;
	statusDistribution: Record<string, number>;
	statusDistributionType: Record<string, Record<string, number>>;
	ratingStats: RatingStats;
	activityByMonth: ActivityPeriod[];
	activityByYear: ActivityPeriod[];
	genreDistribution: GenreStat[];
	recentActivity: RecentActivityItem[];
	hasPartialHistory: boolean;
}

export interface UserSettings {
	private: boolean;
	privateThoughts: boolean;
	hideSpoilers: boolean;
	includePreviouslyWatched: boolean;
	country: string;
	language?: string;
	automateShowStatuses: boolean;
	calendarWebhookUrl?: string;
	calendarWebhookEnabled?: boolean;
	calendarWebhookFilter?: "all" | "movie" | "tv";
	activityWebhookUrl?: string;
	activityWebhookEnabled?: boolean;
	activityWebhookFilter?: "all" | "finished" | "watching" | "rating";
	ratingSystem?: RatingSystem;
	/**
	 * A rating step decided by the user, only
	 * applicable for OutOf10 and OutOf5 rating systems.
	 * Supported: 1, 0.5, 0.1 (must validate).
	 */
	ratingStep?: RatingStep;
}

export enum RatingSystem {
	OutOf10, // default
	OutOf100,
	OutOf5,
	Thumbs,
}

export enum RatingStep {
	One, // default
	Point5,
	Point1,
}

export interface ChangePasswordForm {
	currentPassword: string;
	newPassword: string;
	reEnteredNewPassword: string;
}

// What the user search returns
export interface PublicUser {
	id: number;
	username: string;
	avatar?: Image;
	bio?: string;
}

// PrivateUser - Current users info
export interface PrivateUser {
	username: string;
	type: UserType;
	permissions: UserPermission;
	avatar: Image;
	bio: string;
}

export enum UserPermission {
	PERM_NONE = 1,
	PERM_ADMIN = 2,
	PERM_REQUEST_CONTENT = 4,
	PERM_REQUEST_CONTENT_AUTO_APPROVE = 8,
}

export interface Image {
	createdAt: Date;
	blurHash: string;
	path: string;
}

export interface JellyfinFoundContent {
	hasContent: boolean;
	url: string;
}

export interface AvailableAuthProviders {
	available: string[];
	signupEnabled: boolean;
	isInSetup: boolean;
	useEmby: boolean;
	headerAuthAutoLogin: boolean;
}

export interface TokenClaims {
	userId: number;
	username: string;
	type: number;
}

export interface MediaIDs {
	tmdb?: number;
	imdb?: string;
	wikidata?: string;
	tvdb?: number;

	igdb?: number;
}

export enum MediaTypeE {
	tmdbMovie = "tmdb_movie",
	tmdbShow = "tmdb_tv",
	tmdbPerson = "tmdb_person",
	igdbGame = "igdb_game",
}

export type CancellationStatus =
	| "cliffhanger"
	| "canceled"
	| "hiatus"
	| "concluded"
	| "returning"
	| "unknown";

export interface CancellationInfo {
	status: CancellationStatus;
	badgeText: string;
	description: string;
	warning?: string;
	hasCliffhanger: boolean;
	yearsSinceAir?: number;
}

export interface MediaNetwork {
	id?: number;
	name: string;
	logoPath?: string;
	originCountry?: string;
}

export interface MediaCreator {
	id?: number;
	name: string;
	profilePath?: string;
}

export interface MediaEpisodeSummary {
	seasonNumber: number;
	episodeNumber: number;
	name?: string;
	airDate?: string;
	runtime?: number;
	overview?: string;
}

export interface MediaTimeToBeat {
	hastily?: number;
	normally?: number;
	completely?: number;
	count?: number;
}

export interface ExternalRatings {
	imdbRating?: string;
	imdbVotes?: string;
	rottenTomatoes?: string;
	rottenTomatoesAudience?: string;
	metacritic?: string;
	letterboxd?: string;
	trakt?: string;
	tvmaze?: string;
	anilist?: string;
	malId?: number;
	animationStudio?: string;
	sourceMaterial?: string;
	awards?: string;
	boxOffice?: string;
	networkSchedule?: string;
}

export interface Media {
	type?: MediaTypeE;
	ids: MediaIDs;
	name?: string;
	summary?: string;
	tagline?: string;
	originalTitle?: string;
	originalLanguage?: string;
	certification?: string;
	poster?: Image;
	extPosterPath?: string;
	rating?: number;
	ratingCount?: number;
	watched?: Watched;
	similar?: Media[];
	recommendations?: Media[];
	releaseDate?: string;
	releaseDateLast?: string;
	status?: string;
	extBackdropPath?: string;
	genres?: MediaGenre[];
	homepage?: string;
	videos?: MediaVideo[];
	runtime?: number;
	budget?: number;
	revenue?: number;
	numberOfEpisodes?: number;
	numberOfSeasons?: number;
	networks?: MediaNetwork[];
	productionCompanies?: MediaNetwork[];
	createdBy?: MediaCreator[];
	lastEpisode?: MediaEpisodeSummary;
	nextEpisode?: MediaEpisodeSummary;
	externalRatings?: ExternalRatings;
	providers?: MediaProvider[];
	providersFullListLink?: string;
	gameModes?: MediaGenre[];
	themes?: MediaGenre[];
	storyline?: string;
	criticRating?: number;
	criticRatingCount?: number;
	platforms?: string[];
	developers?: string[];
	publishers?: string[];
	playerPerspectives?: string[];
	screenshots?: string[];
	artworks?: string[];
	timeToBeat?: MediaTimeToBeat;
	seasons?: MediaSeason[];
	isShowAnime?: boolean;
	collection?: MediaCollection;
	cancellation?: CancellationInfo;
}

export interface MediaCollection {
	id: number;
	name: string;
	overview?: string;
	posterPath?: string;
	backdropPath?: string;
	parts?: Media[];
}

export function getContentTypeFromMedia(m: Media): ContentType | undefined {
	switch (m.type) {
		case MediaTypeE.tmdbMovie:
			return "movie";
		case MediaTypeE.tmdbShow:
			return "tv";
		case MediaTypeE.igdbGame:
			return "game";
	}
	return;
}

export interface MediaGenre {
	// ID Is for external db id.
	id: number;
	// Genre name.
	name: string;
}

export enum MediaProviderType {
	Sub = "sub",
	Free = "free",
	Ads = "ads",
	Rent = "rent",
	Buy = "buy",
}

export interface MediaProvider {
	name: string;
	type?: MediaProviderType;
	link?: string;
	logoPath?: string;
}

export enum MediaVideoType {
	trailer = "trailer",
	teaser = "teaser",
	clip = "clip",
	featurette = "featurette",
	behind_the_scenes = "behind_the_scenes",
	bloopers = "bloopers",
	opening_credits = "opening_credits",
	other = "other",
}

export interface MediaVideo {
	id?: string;
	name?: string;
	type?: MediaVideoType;
	best?: boolean;
}

export interface MediaSeason {
	name?: string;
	number: number;
	episodeCount: number;
	releaseDate?: string;
}

interface PaginationParams {
	limit?: number;
	page?: number;
}

export enum SearchType {
	multi = "multi",
	movie = "movie",
	show = "show",
	person = "person",
	game = "game",
}

export interface SearchRequest extends PaginationParams {
	type?: SearchType;
	query: string;
	preferMyList?: boolean;
}

export interface SearchResponseMeta {
	fromMyList?: boolean;
}

export enum DiscoverFilter {
	trending = "trending",
	popular = "popular",
	upcoming = "upcoming",
	streaming = "streaming",
	inTheatres = "intheatres",
}

export type DiscoverFilterOption = `${DiscoverFilter}`;

export interface DiscoverRequest extends PaginationParams {
	type?: SearchType;
	filter?: DiscoverFilter;
}

export interface PersonDetailsResponse {
	name?: string;
	birthday?: string;
	deathday?: string;
	age?: number;
	placeOfBirth?: string;
	knownForDepartment?: string;
	biography?: string;
	homepage?: string;
	extPosterPath?: string;
}

export interface PersonCreditsResponse {
	credits?: Media[];
}

export interface TMDBSeasonDetails {
	_id: string;
	air_date: string;
	episodes: TMDBSeasonDetailsEpisode[];
	name: string;
	overview: string;
	id: number;
	poster_path: string;
	season_number: number;
}

export interface TMDBSeasonDetailsEpisode {
	air_date: string;
	episode_number: number;
	id: number;
	name: string;
	overview: string;
	production_code: string;
	runtime: number;
	season_number: number;
	show_id: number;
	still_path: string;
	vote_average: number;
	vote_count: number;
	crew: {
		department: string;
		job: string;
		credit_id: string;
		adult: boolean;
		gender: number;
		id: number;
		known_for_department: string;
		name: string;
		original_name: string;
		popularity: number;
		profile_path: string;
	}[];
	guest_stars: {
		character: string;
		credit_id: string;
		order: number;
		adult: boolean;
		gender: number;
		id: number;
		known_for_department: string;
		name: string;
		original_name: string;
		popularity: number;
		profile_path: string;
	}[];
}

export interface TMDBContentCredits {
	id: number;
	cast: {
		adult: boolean;
		gender: number;
		id: number;
		known_for_department: string;
		name: string;
		original_name: string;
		popularity: number;
		profile_path: string;
		cast_id: number;
		character: string;
		credit_id: string;
		order: number;
	}[];
	crew: TMDBContentCreditsCrew[];
}

export interface TMDBContentCreditsCrew {
	adult: boolean;
	gender: number;
	id: number;
	known_for_department: string;
	name: string;
	original_name: string;
	popularity: number;
	profile_path: string;
	credit_id: string;
	department: string;
	job: string;
}

export interface TMDBRegions {
	results: {
		iso_3166_1: string;
		english_name: string;
		native_name: string;
	}[];
}

export enum ImportResponseType {
	IMPORT_SUCCESS = "IMPORT_SUCCESS",
	IMPORT_FAILED = "IMPORT_FAILED",
	IMPORT_MULTI = "IMPORT_MULTI",
	IMPORT_NOTFOUND = "IMPORT_NOTFOUND",
	IMPORT_EXISTS = "IMPORT_EXISTS",
}

export interface ImportResponse {
	type: ImportResponseType;
	results?: Media[];
	match?: Media;
	watchedEntry?: Watched;
}

export interface ImportedList {
	tmdbId?: number;
	imdbId?: string;
	igdbId?: number;

	name?: string;
	year?: number;
	type?: "movie" | "tv" | "tv_episode" | "game";
	state?: ImportResponseType;
	rating?: number;
	ratingCustomDate?: Date;
	status?: WatchedStatus;
	thoughts?: string;
	datesWatched?: Date[];
	activity?: Activity[];
	watchedEpisodes?: WatchedEpisode[];
	watchedSeasons?: WatchedSeason[];
	tags?: TagAddRequest[];
}

export interface YamtrackRow {
	media_id?: string | number;
	source?: string;
	media_type?: string;
	title?: string;
	image?: string;
	season_number?: string | number;
	episode_number?: string | number;
	score?: string | number;
	status?: string;
	notes?: string;
	start_date?: string;
	end_date?: string;
	progress?: string | number;
	created_at?: string;
	progressed_at?: string;
}

export interface LetterboxdRow {
	Date?: string;
	Name?: string;
	Year?: string | number;
	"Letterboxd URI"?: string;
	Rating?: string | number;
	Rewatch?: string;
	Tags?: string;
	"Watched Date"?: string;
	Description?: string;
}

export interface Filters {
	type: string[];
	status: string[];
	year?: number;
	minRating?: number;
	maxRating?: number;
	query?: string;
	genre?: string;
	productionStatus?: string;
}

export interface ManagedUser {
	id: number;
	createdAt: Date;
	username: string;
	type: UserType;
	permissions: number;
	private: boolean;
}

export interface ServerConfig {
	DEFAULT_COUNTRY?: string;
	DEFAULT_LANGUAGE?: string;
	JELLYFIN_HOST?: string;
	JELLYFIN_WEBHOOK_ENABLED?: boolean;
	JELLYFIN_WEBHOOK_KEY?: string;
	USE_EMBY: boolean;
	SIGNUP_ENABLED: boolean;
	TMDB_KEY?: string;
	OMDB_KEY?: string;
	MDBLIST_KEY?: string;
	PLEX_HOST?: string;
	PLEX_MACHINE_ID?: string;
	OVERSEERR_HOST?: string;
	OVERSEERR_KEY?: string;
	SONARR?: SonarrSettings[];
	RADARR?: RadarrSettings[];
	TWITCH?: TwitchSettings;
	DEBUG: boolean;
}

export interface ServerConfigByName<T> {
	value: T;
}

export interface SonarrSettings {
	name: string;
	host?: string;
	key?: string;
	qualityProfile?: number;
	rootFolder?: number;
	languageProfile?: number;
	automaticSearch?: boolean;
}

export interface RadarrSettings {
	name: string;
	host?: string;
	key?: string;
	qualityProfile?: number;
	rootFolder?: number;
	automaticSearch?: boolean;
}

interface ArrSettingsPublicResponseBase {
	name: string;
	host?: string;
	qualityProfile?: number;
	rootFolder?: number;
	automaticSearch: boolean;
}

export interface SonarrSettingsPublicResponseResult extends ArrSettingsPublicResponseBase {
	languageProfile?: number;
}

export interface RadarrSettingsPublicResponseResult extends ArrSettingsPublicResponseBase {}

export interface TwitchSettings {
	clientId?: string;
	clientSecret?: string;
}

export interface TrustedHeaderAuthSetting {
	enabled: boolean;
	headerName: string;
	autoLogin: boolean;
	logoutUrl: string;
}

export interface TrustedHeaderAuthLogoutDetailsResponse {
	logoutUrl?: string;
}

export interface DropDownItem {
	id: number | string;
	value: string;
	icon?: Icon;
}

export interface ListBoxItem {
	id: number;
	value: boolean;
	displayValue: string;
}

export interface QualityProfile {
	name: string;
	upgradeAllowed: boolean;
	cutoff: number;
	items: {
		quality?: {
			id: number;
			name: string;
			source: string;
			resolution: number;
		};
		items: unknown[];
		allowed: boolean;
		name?: string;
		id?: number;
	}[];
	id: number;
}

export interface RootFolder {
	path: string;
	accessible: boolean;
	freeSpace: number;
	unmappedFolders: unknown[];
	id: number;
}

export interface LanguageProfile {
	name: string;
	upgradeAllowed: boolean;
	cutoff: {
		id: number;
		name: string;
	};
	languages: {
		language: {
			id: number;
			name: string;
		};
		allowed: boolean;
	}[];
	id: number;
}

export interface SonarrTestResponse {
	qualityProfiles: QualityProfile[];
	rootFolders: RootFolder[];
	languageProfiles: LanguageProfile[];
}

export interface RadarrTestResponse {
	qualityProfiles: QualityProfile[];
	rootFolders: RootFolder[];
}

export type ArrRequestStatus =
	"PENDING" | "APPROVED" | "AUTO_APPROVED" | "DENIED" | "FOUND";

export interface ArrRequestResponse {
	id: number;
	createdAt: string;
	updatedAt: string;
	serverName: string;
	arrId: number;
	content: Content;
	status: ArrRequestStatus;
	requestJson: string;
	username: string;
}

export interface ArrDetailsResponse {
	progress: number;
	estimatedCompletionTime: string;
	status: string;
	trackedDownloadStatus: string;
	trackedDownloadState: string;
}

export interface ArrInfoResponse {
	hasFile: boolean;
	isAvailable: boolean;
	sizeOnDisk?: number;
	qualityName?: string;
	added: string;
	statistics?: {
		seasonCount: number;
		episodeFileCount: number;
		episodeCount: number;
		totalEpisodeCount: number;
		sizeOnDisk: number;
		percentOfEpisodes: number;
	};
}

export interface ServerFeatures {
	sonarr: boolean;
	radarr: boolean;
	games: boolean;
	overseerr?: boolean;
}

export interface Follow {
	createdAt: Date;
	followedUser: PublicUser;
}

interface MovaryExportBase {
	title: string;
	year: string;
	tmdbId: string;
	imdbId: string;
}

export interface MovaryHistory extends MovaryExportBase {
	watchedAt: string;
	comment: string;
}

export interface MovaryRatings extends MovaryExportBase {
	userRating: string;
}

export interface MovaryWatchlist extends MovaryExportBase {
	addedAt: string;
}

export interface TodoMoviesExport {
	Movie: TodoMoviesMovie[];
	MovieList: TodoMoviesCustomList[];
}

export interface TodoMoviesMovie {
	Attrs: {
		tmdbID: number;
		title: string;
		isWatched: number;
		insertionDate: {
			Value: number;
			Class: string;
		};
		myScore: number;
	};
	Rels: {
		lists: {
			Items: string[];
			Entity: string;
		};
	};
	ObjectID: string;
}

export interface TodoMoviesCustomList {
	Attrs: {
		colorInHex: string;
		order: number;
		iconFileName: string;
		featuredListID: number;
		name: string;
	};
	Rels: {
		movies: {
			Items: string[];
			Entity: string;
		};
	};
	ObjectID: string;
}

// General interface for all requests that return a job that was started.
export interface JobCreatedResponse {
	jobId: string;
}

export enum JobStatus {
	CREATED = "CREATED",
	RUNNING = "RUNNING",
	DONE = "DONE",
	CANCELLED = "CANCELLED",
}

export interface GetJobResponse {
	name: string;
	status: JobStatus;
	currentTask?: string;
	errors: string[];
}

export interface TaskRescheduleRequest {
	seconds: number;
}

export interface AllTasksResponse {
	name: string;
	nextRun: Date;
	seconds: number;
}

export interface Tag extends dbModel {
	name: string;
	color: string;
	bgColor: string;
	description?: string;
	isPublic?: boolean;
}

export interface TagAddRequest {
	name: string;
	color: string;
	bgColor: string;
	description?: string;
	isPublic?: boolean;
}

export interface PublicTagResponse {
	id: number;
	name: string;
	color: string;
	bgColor: string;
	description?: string;
	isPublic: boolean;
	authorName: string;
	totalItems: number;
	items: Media[];
}

export type CalendarFilter = "library" | "all";
export type CalendarType = "all" | "movie" | "tv";

export interface CalendarEvent {
	id: string;
	mediaId: number;
	mediaType: "movie" | "tv";
	title: string;
	releaseDate: string; // YYYY-MM-DD
	posterPath?: string;
	backdropPath?: string;
	overview?: string;
	inLibrary: boolean;
	status?: WatchedStatus;
	seasonNumber?: number;
	episodeNumber?: number;
	episodeName?: string;
}

export interface FranchiseItem {
	tmdbId: number;
	type: "movie" | "tv";
	title: string;
	year: string;
	posterPath?: string;
	backdropPath?: string;
	overview?: string;
	phaseOrEra?: string;
	timelineOrder: number;
	releaseOrder: number;
	releaseDate?: string;
	watched?: Watched;
	isWatched: boolean;
	isNextUp: boolean;
}

export interface FranchiseRoadmap {
	id: string;
	name: string;
	description: string;
	backdropPath?: string;
	posterPath?: string;
	activeOrder: "timeline" | "release";
	totalItems: number;
	watchedCount: number;
	percentWatched: number;
	nextUpItem?: FranchiseItem;
	items: FranchiseItem[];
}

export interface FranchiseSummary {
	id: string;
	name: string;
	description: string;
	backdropPath?: string;
	posterPath?: string;
	totalItems: number;
	watchedCount: number;
	percentWatched: number;
	itemTypes: string;
}

export interface FranchiseMatch {
	inFranchise: boolean;
	franchiseId?: string;
	franchiseName?: string;
	phaseOrEra?: string;
	timelineOrder?: number;
	releaseOrder?: number;
}


export interface SeerrContentStatus {
	configured: boolean;
	status: number;
	statusName: "UNKNOWN" | "PENDING" | "PROCESSING" | "PARTIALLY_AVAILABLE" | "AVAILABLE";
	isAvailable: boolean;
	isProcessing: boolean;
	isPending: boolean;
	mediaInfo?: {
		id?: number;
		tmdbId?: number;
		tvdbId?: number;
		status: number;
	};
}


export interface WatchTogetherUser {
	id: number;
	username: string;
	bio?: string;
	avatar?: Image;
	isFollowed: boolean;
}

export interface WatchTogetherItem {
	id: number;
	contentId: number;
	status: WatchedStatus;
	rating: number;
	content?: Content;
	targetUserStatus?: WatchedStatus;
	targetUserRating?: number;
}

export interface WatchTogetherResponse {
	targetUser: PublicUser;
	mode: "common" | "unwatched" | "recommended";
	totalCount: number;
	items: WatchTogetherItem[];
}

export interface BackupInfo {
	filename: string;
	size: number;
	formattedSize: string;
	createdAt: string;
	isAuto: boolean;
	type: "sqlite" | "json";
}

export interface BackupConfig {
	enabled: boolean;
	retention: number;
	scheduleSeconds: number;
	nextRun?: string;
	totalBackups: number;
	totalSize: number;
	formattedTotalSize: string;
	driver: string;
	lastBackup?: string;
}

export interface GetBackupsResponse {
	backups: BackupInfo[];
	config: BackupConfig;
}

export interface UpdateBackupConfigRequest {
	enabled?: boolean;
	retention?: number;
	scheduleSeconds?: number;
}
