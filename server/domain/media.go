// Types that we can use for all content types (movie, tv, game, everything).
// Data responses to the client can use these "uniform" types to make access
// easier.

package domain

import (
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/util"
)

type MediaType string

const (
	MediaTypeTMDBMovie  MediaType = "tmdb_movie"
	MediaTypeTMDBShow   MediaType = "tmdb_tv"
	MediaTypeTMDBPerson MediaType = "tmdb_person"

	MediaTypeIGDBGame MediaType = "igdb_game"
)

type CancellationStatus string

const (
	CancellationStatusCliffhanger CancellationStatus = "cliffhanger"
	CancellationStatusCanceled    CancellationStatus = "canceled"
	CancellationStatusHiatus      CancellationStatus = "hiatus"
	CancellationStatusConcluded   CancellationStatus = "concluded"
	CancellationStatusReturning   CancellationStatus = "returning"
	CancellationStatusUnknown     CancellationStatus = "unknown"
)

type CancellationInfo struct {
	Status         CancellationStatus `json:"status"`
	BadgeText      string             `json:"badgeText"`
	Description    string             `json:"description"`
	Warning        string             `json:"warning,omitempty"`
	HasCliffhanger bool               `json:"hasCliffhanger"`
	YearsSinceAir  float64            `json:"yearsSinceAir,omitempty"`
}

type CuratedCliffhangerEntry struct {
	Title           string
	Seasons         string
	CliffhangerNote string
}

// CuratedCliffhangers maps TMDB Show IDs to known abrupt cancellations with cliffhangers
var CuratedCliffhangers = map[int]CuratedCliffhangerEntry{
	90669:  {Title: "1899", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix com a grande reviravolta da realidade simulada e da nave Prometheus em aberto."},
	67744:  {Title: "Mindhunter", Seasons: "2 temporadas", CliffhangerNote: "Hiato indefinido / cancelada sem desfecho para o desenvolvimento da caça ao assassino BTK."},
	69061:  {Title: "The OA", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Netflix no pico do salto dimensional meta-narrativo da 2ª temporada."},
	153096: {Title: "KAOS", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix logo após a 1ª temporada com a revolta contra Zeus e as profecias em aberto."},
	135157: {Title: "Archive 81", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix na 1ª temporada com Dan preso no hospital em 1994 e mistérios da seita por responder."},
	69407:  {Title: "Santa Clarita Diet", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pela Netflix na 3ª temporada com a mordida em Joel e o parasitismo de Mr. Ball Legs."},
	63247:  {Title: "Westworld", Seasons: "4 temporadas", CliffhangerNote: "Cancelada pela HBO antes da 5ª temporada planeada por Jonathan Nolan e Lisa Joy para concluir o ciclo."},
	85723:  {Title: "Raised by Wolves", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Max na 2ª temporada com mistérios da árvore, da serpente e da entidade sem resposta."},
	4663:   {Title: "Deadwood", Seasons: "3 temporadas", CliffhangerNote: "Cancelada abruptamente pela HBO após 3 temporadas (filme posterior tentou reunir pontas soltas anos depois)."},
	1437:   {Title: "Firefly", Seasons: "1 temporada", CliffhangerNote: "Cancelada prematuramente pela Fox com múltiplos arcos dos tripulantes da Serenity inacabados."},
	2352:   {Title: "My Name Is Earl", Seasons: "4 temporadas", CliffhangerNote: "Cancelada pela NBC após a 4ª temporada terminando com um 'To Be Continued...' e a paternidade de Earl Jr. suspensa."},
	64196:  {Title: "Colony", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pela USA Network na 3ª temporada segundos antes do ataque alienígena frontal."},
	20159:  {Title: "FlashForward", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela ABC na 1ª temporada durante o segundo grande apagão mundial e fuga do edifício."},
	1398:   {Title: "Jericho", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela CBS com a iminência da segunda guerra civil entre Cheyenne e Columbus."},
	51019:  {Title: "Almost Human", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Fox após 1 temporada deixando o segredo da grande muralha e da robótica sem resposta."},
	104:    {Title: "Carnivàle", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela HBO na 2ª temporada de um plano inicial de seis temporadas da batalha Avatarica."},
	34524:  {Title: "Terra Nova", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Fox na 1ª temporada com a descoberta da proa de navio do século XVIII em pleno Cretáceo."},
	46648:  {Title: "Hannibal", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pela NBC após 3 temporadas terminando com a queda mútua de Will Graham e Hannibal do penhasco."},
	70785:  {Title: "Anne with an E", Seasons: "3 temporadas", CliffhangerNote: "Cancelada abruptamente pela CBC/Netflix ao fim de 3 temporadas sem adaptar a vida adulta dos livros."},
	85720:  {Title: "Shadow and Bone", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Netflix na 2ª temporada com a ameaça de jurda parem e a coroação em aberto."},
	87108:  {Title: "Warrior Nun", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Netflix na 2ª temporada com a passagem de Ava para o portal de Reya e guerra iminente."},
	72758:  {Title: "GLOW", Seasons: "3 temporadas", CliffhangerNote: "Renovada para a 4ª temporada e revertida/cancelada durante a pandemia sem filmar o desfecho."},
	88040:  {Title: "The Society", Seasons: "1 temporada", CliffhangerNote: "Renovada para a 2ª temporada e cancelada de seguida devido à COVID-19 com o mistério de New Ham em aberto."},
	96162:  {Title: "Spinning Out", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix após 1 temporada antes da competição final de patinagem artística."},
	97951:  {Title: "Julie and the Phantoms", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix na 1ª temporada com Caleb a possuir Nick."},
	86831:  {Title: "The Order", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Netflix na 2ª temporada com Alyssa ferida mortalmente e Jack com o Vade Maecum."},
	106308: {Title: "Teenage Bounty Hunters", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix na 1ª temporada com a revelação surpresa de que as gémeas são meias-irmãs."},
	98112:  {Title: "I Am Not Okay with This", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix após a 1ª temporada com a figura sombria a aproximar-se de Syd."},
	60708:  {Title: "Constantine", Seasons: "1 temporada", CliffhangerNote: "Cancelada após 1 temporada com a revelação de que Manny liderava a Treva Crescente."},
	61222:  {Title: "Forever", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela ABC na 1ª temporada antes de Henry conseguir revelar a sua imortalidade."},
	3222:   {Title: "Pushing Daisies", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela ABC na 2ª temporada com o encerramento apressado dos arcos principais."},
	39852:  {Title: "Alphas", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pelo Syfy na 2ª temporada com todas as pessoas na Grand Central desmaiadas exceto Gary."},
	1434:   {Title: "Stargate Universe", Seasons: "2 temporadas", CliffhangerNote: "Cancelada na 2ª temporada com a Destiny à deriva entre galáxias e Eli com poucas semanas de suporte de vida."},
	433:    {Title: "Terminator: The Sarah Connor Chronicles", Seasons: "2 temporadas", CliffhangerNote: "Cancelada na 2ª temporada com John a saltar para o futuro onde ninguém ouviu falar dele."},
	61550:  {Title: "Marvel's Agent Carter", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela ABC na 2ª temporada com o assassinato misterioso de Jack Thompson."},
	4087:   {Title: "Kyle XY", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pela ABC Family na 3ª temporada com Cassidy a revelar que era irmão de Kyle."},
	1572:   {Title: "Surface", Seasons: "1 temporada", CliffhangerNote: "Cancelada após 1 temporada com um tsunami massivo e as criaturas marinhas a assumirem a costa."},
	1156:   {Title: "Invasion (2005)", Seasons: "1 temporada", CliffhangerNote: "Cancelada na 1ª temporada com a transformação da esposa de Russell pelas criaturas híbridas."},
	16297:  {Title: "V (2009)", Seasons: "2 temporadas", CliffhangerNote: "Cancelada após 2 temporadas com o envio da Bliss sobre todo o planeta Terra."},
	45045:  {Title: "Revolution", Seasons: "2 temporadas", CliffhangerNote: "Cancelada após 2 temporadas com os nanorobôs a ganharem consciência e assumirem o controlo."},
	62425:  {Title: "Dark Matter", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pelo Syfy na 3ª temporada com a invasão iminente da misteriosa Armada Negra."},
	67136:  {Title: "The Exorcist", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Fox na 2ª temporada deixando o destino trágico do Padre Marcus em aberto."},
	37680:  {Title: "The Borgias", Seasons: "3 temporadas", CliffhangerNote: "Cancelada pela Showtime na 3ª temporada antes do quarto ano planeado por Neil Jordan."},
	124800: {Title: "Our Flag Means Death", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Max antes da 3ª temporada planeada como desfecho da relação e jornada pirata."},
	94664:  {Title: "Lockwood & Co.", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix na 1ª temporada antes de Lockwood revelar o que havia atrás da porta misteriosa."},
	95557:  {Title: "Fate: The Winx Saga", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Netflix na 2ª temporada com Bloom presa no Reino da Escuridão diante da sua mãe."},
	200875: {Title: "Dead Boy Detectives", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix após 1 temporada sem continuação para o universo de Sandman."},
	114479: {Title: "The Acolyte", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Disney+ na 1ª temporada com teasers de Darth Plagueis, Mestre Yoda e a busca de Osha."},
	80885:  {Title: "The Nevers", Seasons: "1 temporada", CliffhangerNote: "Cancelada e removida de catálogo pela Warner Bros Discovery sem desfecho da invasão temporal."},
	93405:  {Title: "Halo", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela Paramount+ na 2ª temporada após a chegada ao anel Halo e a libertação do Flood."},
	204541: {Title: "Scavengers Reign", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Max/Netflix na 1ª temporada com pontas soltas sobre a colonização do ecossistema."},
	119054: {Title: "Inside Job", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela Netflix após renovação anterior, deixando a grande conspiração dos Mantos Negros por resolver."},
	74387:  {Title: "Final Space", Seasons: "3 temporadas", CliffhangerNote: "Cancelada após 3 temporadas com a vitória aparente de Invictus e retirada de circulação como dedução fiscal."},
	46511:  {Title: "Utopia (UK)", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pelo Channel 4 na 2ª temporada com a conspiração Janus em pleno curso."},
	2270:   {Title: "Las Vegas", Seasons: "5 temporadas", CliffhangerNote: "Cancelada na 5ª temporada no meio do casamento de Delinda e o mistério de Cooper."},
	2204:   {Title: "Lois & Clark: The New Adventures of Superman", Seasons: "4 temporadas", CliffhangerNote: "Cancelada na 4ª temporada deixando um bebé misterioso com o símbolo de Krypton no berço."},
	1475:   {Title: "Tru Calling", Seasons: "2 temporadas", CliffhangerNote: "Cancelada a meio da 2ª temporada com a batalha do destino entre Tru e Jack sem conclusão."},
	2697:   {Title: "John Doe", Seasons: "1 temporada", CliffhangerNote: "Cancelada após 1 temporada revelando que o amigo mais próximo fazia parte da conspiração."},
	153496: {Title: "Quantum Leap (2022)", Seasons: "2 temporadas", CliffhangerNote: "Cancelada pela NBC após 2 temporadas com Ben e Hannah presos no tempo."},
	104877: {Title: "Debris", Seasons: "1 temporada", CliffhangerNote: "Cancelada pela NBC após 1 temporada com anomalias de clones e realidades convergentes em aberto."},
}

// GetCliffhangerInfo returns CancellationInfo if the tmdb ID is in CuratedCliffhangers
func GetCliffhangerInfo(tmdbID int) *CancellationInfo {
	if entry, ok := CuratedCliffhangers[tmdbID]; ok {
		return &CancellationInfo{
			Status:         CancellationStatusCliffhanger,
			BadgeText:      "Cancelada sem Desfecho",
			Description:    fmt.Sprintf("Cancelada abruptamente (%s): %s", entry.Seasons, entry.CliffhangerNote),
			Warning:        fmt.Sprintf("Atenção: Esta série foi cancelada após %s e termina com um gancho em aberto (cliffhanger).", entry.Seasons),
			HasCliffhanger: true,
		}
	}
	return nil
}

type Media struct {
	// The type of media.
	Type MediaType `json:"type,omitempty"`
	// The ids associated with this media.
	IDs MediaIDs `json:"ids"`
	// The name of the media.
	Name string `json:"name,omitempty"`
	// A description.
	Summary string `json:"summary,omitempty"`
	// The poster.
	Poster *entity.Image `json:"poster,omitempty"`
	// The external poster path.
	ExtPosterPath string `json:"extPosterPath,omitempty"`
	// The rating.
	Rating uint `json:"rating,omitempty"`
	// The amount of votes that made up the rating.
	RatingCount uint `json:"ratingCount,omitempty"`
	// Watched data.
	Watched WatchedDto `json:"watched,omitzero"`
	// Similar media.
	Similar []Media `json:"similar,omitempty"`
	// Release date / first air date.
	ReleaseDate time.Time `json:"releaseDate,omitzero"`
	// Videos (trailers, etc)
	Videos []MediaVideo `json:"videos,omitempty"`
	// Media collection / series (for movies).
	Collection *MediaCollection `json:"collection,omitempty"`

	//
	// Properties that are less important (not used for all responses).
	//

	// Backdrop path.
	ExtBackdropPath string `json:"extBackdropPath,omitempty"`
	// Genres.
	Genres []MediaGenre `json:"genres,omitempty"`
	// Media website.
	Homepage string `json:"homepage,omitempty"`
	// Media providers (eg Streaming sites, game markets)
	Providers []MediaProvider `json:"providers,omitempty"`
	// A link to the database we are using that lists all providers with max details.
	// (especially for TMDB since it's data from JustWatch isn't available to us).
	ProvidersFullListLink string `json:"providersFullListLink,omitempty"`
	// Status of the media (released, ended, etc).
	// Depending on media type, this will contain different values.
	Status string `json:"status,omitempty"`

	//
	// Properties only for movies/tv.
	//

	// Runtime.
	Runtime uint `json:"runtime,omitempty"`
	// Seasons.
	Seasons []MediaSeason `json:"seasons,omitempty"`
	// Simple bool for our RequestShow component since Sonarr can be given a
	// series type (ideally the frontend doesn't need to do that, but for now it
	// does.. if (son)arr code is refactored, can the client just pass very basic
	// details for the server to fetch fully/verify, i.e fetched full details from
	// tmdb again to verify if show is anime itself, etc).
	IsShowAnime bool `json:"isShowAnime,omitempty"`
	// Last release date.
	// Currently used for tv shows so frontend can display its end date.
	ReleaseDateLast time.Time `json:"releaseDateLast,omitzero"`
	// Cancellation Guard metadata for TV shows (cliffhangers, hiatus, conclusion)
	Cancellation *CancellationInfo `json:"cancellation,omitempty"`

	//
	// Properties only for Games
	//

	// Game modes.
	GameModes []MediaGenre `json:"gameModes,omitempty"`
}

type MediaCollection struct {
	ID           int     `json:"id"`
	Name         string  `json:"name"`
	Overview     string  `json:"overview,omitempty"`
	PosterPath   string  `json:"posterPath,omitempty"`
	BackdropPath string  `json:"backdropPath,omitempty"`
	Parts        []Media `json:"parts,omitempty"`
}

func (t Media) GetId() int {
	switch t.Type {
	case MediaTypeTMDBMovie,
		MediaTypeTMDBShow:
		return t.IDs.TMDB
	case MediaTypeIGDBGame:
		return t.IDs.IGDB
	}
	return -99
}

// If this changes, verify all use cases still make sense!
func (t Media) GetMediaType() util.SupportedMedia {
	switch t.Type {
	case MediaTypeTMDBMovie:
		return util.SupportedMediaMovie
	case MediaTypeTMDBShow:
		return util.SupportedMediaShow
	case MediaTypeIGDBGame:
		return util.SupportedMediaGame
	}
	// Unsupported...
	slog.Warn("GetMediaType: Requested, but unsupported type encountered.",
		"type", t.Type)
	return ""
}

type MediaIDs struct {
	// The internal ID
	// Watcharr uint

	// For tmdb data
	TMDB     int    `json:"tmdb,omitempty"`
	IMDB     string `json:"imdb,omitempty"`
	Wikidata string `json:"wikidata,omitempty"`
	TVDB     int    `json:"tvdb,omitempty"`

	// For igdb data
	IGDB int `json:"igdb,omitempty"`
}

type MediaGenre struct {
	// ID of the genre on the external database.
	ID uint `json:"id,omitempty"`
	// Name of genre.
	Name string `json:"name,omitempty"`
}

type MediaSeason struct {
	// Season number (doesn't omit empty to keep support for season 0).
	Number int `json:"number"`
	// Season name.
	Name string `json:"name,omitempty"`
	// Season air date.
	ReleaseDate time.Time `json:"releaseDate,omitzero"`
	// Number of episodes in season.
	EpisodeCount int `json:"episodeCount"`
}

// Create Media dto from Watched entity.
func NewMediaFromWatched(w *entity.Watched, watchedDto *WatchedDto) Media {
	var media Media

	if w.Content != nil {
		media = NewMediaFromContent(w.Content)
	} else if w.Game != nil {
		media = NewMediaFromGame(w.Game)
	}

	media.Watched = *watchedDto

	return media
}

// Converter for Content (tv/movie) entity to Media
func NewMediaFromContent(c *entity.Content) Media {
	m := Media{
		IDs: MediaIDs{
			TMDB: c.TmdbID,
		},
		Name:          c.Title,
		Summary:       c.Overview,
		ExtPosterPath: c.PosterPath,
		Rating:        uint(math.Round(float64(c.VoteAverage * 10))),
		RatingCount:   uint(c.VoteCount),
		Runtime:       uint(c.Runtime),
		Status:        c.Status,
	}
	switch c.Type {
	case entity.MOVIE:
		m.Type = MediaTypeTMDBMovie
	case entity.SHOW:
		m.Type = MediaTypeTMDBShow
	}
	if c.ReleaseDate != nil {
		m.ReleaseDate = *c.ReleaseDate
	}
	return m
}

// Converter for Game entity to Media
func NewMediaFromGame(c *entity.Game) Media {
	m := Media{
		IDs: MediaIDs{
			IGDB: c.IgdbID,
		},
		Type:          MediaTypeIGDBGame,
		Name:          c.Name,
		Summary:       c.Summary,
		Poster:        c.Poster,
		ExtPosterPath: c.CoverID,
		Rating:        uint(c.Rating),
		RatingCount:   uint(c.RatingCount),
	}
	if c.ReleaseDate != nil {
		m.ReleaseDate = *c.ReleaseDate
	}
	return m
}
