package content

import (
	"fmt"
	"strings"
	"time"

	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/media/tmdb"
)

// EvaluateShowCancellation computes the CancellationInfo for a TV show combining
// community overrides, curated cliffhanger lists, and TMDB status heuristics.
func EvaluateShowCancellation(
	show tmdb.ShowDetails,
	dbOverride string,
) domain.CancellationInfo {
	// 1. Check database override first if present
	switch strings.ToLower(strings.TrimSpace(dbOverride)) {
	case "cliffhanger":
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusCliffhanger,
			BadgeText:      "Cancelada sem Desfecho",
			Description:    "Esta série foi marcada pela comunidade ou pelo utilizador como cancelada abruptamente com pontas soltas (cliffhanger).",
			Warning:        "Atenção: Esta série foi descontinuada e termina com ganchos importantes por resolver.",
			HasCliffhanger: true,
		}
	case "conclusive":
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusConcluded,
			BadgeText:      "História Concluída",
			Description:    "Esta série concluiu o seu arco narrativo de forma conclusiva e planeada.",
			HasCliffhanger: false,
		}
	case "canceled":
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusCanceled,
			BadgeText:      "Cancelada pela Produção",
			Description:    "Esta série foi formalmente cancelada pela emissora ou plataforma.",
			Warning:        "Cancelada pela produção.",
			HasCliffhanger: false,
		}
	case "hiatus":
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusHiatus,
			BadgeText:      "Hiato Prolongado",
			Description:    "Esta série está sem novos episódios há um longo período de tempo.",
			HasCliffhanger: false,
		}
	}

	// 2. Check Curated Cliffhangers dataset
	if info := domain.GetCliffhangerInfo(show.ID); info != nil {
		return *info
	}

	// 3. TMDB Status checks
	statusStr := strings.TrimSpace(string(show.Status))

	// Formally Canceled on TMDB
	if strings.EqualFold(statusStr, "Canceled") {
		var seasonsDesc string
		if show.NumberOfSeasons > 0 {
			if show.NumberOfSeasons == 1 {
				seasonsDesc = "após 1 temporada"
			} else {
				seasonsDesc = fmt.Sprintf("após %d temporadas", show.NumberOfSeasons)
			}
		}
		desc := "A emissora ou plataforma cancelou oficialmente a produção desta série."
		warn := "Atenção: Série cancelada oficialmente pela produção."
		if seasonsDesc != "" {
			desc = fmt.Sprintf("A produção foi cancelada oficialmente %s.", seasonsDesc)
			warn = fmt.Sprintf("Atenção: Esta série foi cancelada %s e poderá não ter um desfecho conclusivo.", seasonsDesc)
		}
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusCanceled,
			BadgeText:      "Cancelada pela Produção",
			Description:    desc,
			Warning:        warn,
			HasCliffhanger: false,
		}
	}

	// Check prolonged hiatus for "Returning Series" or active production
	if strings.EqualFold(statusStr, "Returning Series") || show.InProduction {
		if show.LastAirDate != "" {
			if lastDate, err := time.Parse("2006-01-02", show.LastAirDate); err == nil {
				years := time.Since(lastDate).Hours() / (24 * 365.25)
				// If last air date is > 2.5 years ago and no next episode announced
				if years >= 2.5 && (show.NextEpisodeToAir == nil || show.NextEpisodeToAir.AirDate == "") {
					return domain.CancellationInfo{
						Status:         domain.CancellationStatusHiatus,
						BadgeText:      "Hiato Prolongado (+2 anos)",
						Description:    fmt.Sprintf("O último episódio estreou há %.1f anos (%s) sem novos episódios calendarizados. Existe elevado risco de descontinuação ou cancelamento silencioso.", years, show.LastAirDate),
						Warning:        fmt.Sprintf("Sem novos episódios há %.1f anos. A série poderá nunca ser retomada.", years),
						HasCliffhanger: false,
						YearsSinceAir:  years,
					}
				}
			}
		}
		// Active returning series
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusReturning,
			BadgeText:      "Em Exibição",
			Description:    "Série ativa em exibição ou com novas temporadas planeadas.",
			HasCliffhanger: false,
		}
	}

	// Miniseries or Ended
	if strings.EqualFold(statusStr, "Ended") || strings.EqualFold(show.Type, "Miniseries") {
		desc := "Esta produção concluiu o seu ciclo de exibição."
		if strings.EqualFold(show.Type, "Miniseries") || (show.NumberOfSeasons == 1 && strings.EqualFold(statusStr, "Ended")) {
			desc = "Minissérie com narrativa fechada e ciclo concluído."
		}
		return domain.CancellationInfo{
			Status:         domain.CancellationStatusConcluded,
			BadgeText:      "História Concluída",
			Description:    desc,
			HasCliffhanger: false,
		}
	}

	// Default fallback
	return domain.CancellationInfo{
		Status:         domain.CancellationStatusUnknown,
		BadgeText:      statusStr,
		Description:    statusStr,
		HasCliffhanger: false,
	}
}

