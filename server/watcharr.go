// Watcharr is licensed under the GPLv3 license.

package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime/debug"
	"strconv"
	"strings"
	"syscall"
	"time"

	_ "embed"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/sbondCo/Watcharr/config"
	"github.com/sbondCo/Watcharr/database"
	"github.com/sbondCo/Watcharr/database/entity"
	"github.com/sbondCo/Watcharr/domain"
	"github.com/sbondCo/Watcharr/feature/activity"
	"github.com/sbondCo/Watcharr/feature/backup"
	"github.com/sbondCo/Watcharr/feature/arr"
	"github.com/sbondCo/Watcharr/feature/auth"
	"github.com/sbondCo/Watcharr/feature/calendar"
	"github.com/sbondCo/Watcharr/feature/content"
	"github.com/sbondCo/Watcharr/feature/discover"
	"github.com/sbondCo/Watcharr/feature/feature"
	"github.com/sbondCo/Watcharr/feature/follow"
	"github.com/sbondCo/Watcharr/feature/franchise"
	"github.com/sbondCo/Watcharr/feature/game"
	"github.com/sbondCo/Watcharr/feature/img"
	"github.com/sbondCo/Watcharr/feature/imprt"
	"github.com/sbondCo/Watcharr/feature/integration"
	"github.com/sbondCo/Watcharr/feature/jellyfin"
	"github.com/sbondCo/Watcharr/feature/job"
	"github.com/sbondCo/Watcharr/feature/plex"
	"github.com/sbondCo/Watcharr/feature/push"
	"github.com/sbondCo/Watcharr/feature/profile"
	"github.com/sbondCo/Watcharr/feature/search"
	"github.com/sbondCo/Watcharr/feature/seerr"
	"github.com/sbondCo/Watcharr/feature/server"
	"github.com/sbondCo/Watcharr/feature/setup"
	"github.com/sbondCo/Watcharr/feature/tag"
	"github.com/sbondCo/Watcharr/feature/task"
	"github.com/sbondCo/Watcharr/feature/user"
	"github.com/sbondCo/Watcharr/feature/watched"
	"github.com/sbondCo/Watcharr/feature/watched/episode"
	"github.com/sbondCo/Watcharr/feature/watched/season"
	"github.com/sbondCo/Watcharr/logging"
	"github.com/sbondCo/Watcharr/media/tmdb"
	"github.com/sbondCo/Watcharr/router"
	taskl "github.com/sbondCo/Watcharr/task"
	"gorm.io/gorm/logger"
)

//go:embed VERSION
var version string

func main() {
	startTime := time.Now()

	// Tune Go runtime memory footprint: GC at 60% growth prevents RAM bloat,
	// and soft limit (default 180MB, configurable via WATCHARR_MEMORY_LIMIT_MB) keeps working set tightly managed.
	debug.SetGCPercent(60)
	memLimitMB := int64(180)
	if envMem := os.Getenv("WATCHARR_MEMORY_LIMIT_MB"); envMem != "" {
		if parsed, err := strconv.ParseInt(envMem, 10, 64); err == nil && parsed >= 64 {
			memLimitMB = parsed
		}
	}
	debug.SetMemoryLimit(memLimitMB * 1024 * 1024)

	err := godotenv.Load()
	if err != nil {
		// Do not fail if file does not exist
		if !os.IsNotExist(err) {
			log.Fatal("Failed to load vars from .env file:", err)
		}
	}

	multiw := logging.Setup(path.Join(config.DataPath, "watcharr.log"))
	// Just a nice separator so when inspecting the log file I know when
	// logs are for a new instance.
	fmt.Fprintf(multiw, `
	Stand back. Watcharr is starting.

`)
	slog.Info("Watcharr starting", "version", version)
	fmt.Printf(`
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣀⣀⣀⣀⣀⣀⣀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣠⣴⣶⣿⠿⠛⠛⠛⠻⠿⣿⣿⣿⣿⣿⣶⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⣴⣿⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⢀⣿⣿⣿⣿⣿⣷⣻⠶⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠄⠂⠀⢀⣠⣾⣿⣿⣿⣿⣿⣿⣿⡄⠀⠀⠀⢀⣤⣾⣿⣿⣿⣿⣿⣿⡿⣽⣻⣳⢎⡄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠄⢡⠂⠄⣢⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⣶⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣟⡷⣯⡞⣝⢆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⠀⠁⡐⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣳⣟⡾⣹⢎⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠂⣼⣿⣿⣿⣿⡿⠿⠛⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠉⠉⠛⠻⠿⣿⣿⣿⣿⣿⡿⣾⣝⣧⢻⡜⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢂⠐⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡀⠂⢸⣿⡿⠟⠋⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠻⠿⣿⣳⢯⣞⡳⣎⠅⠀⠀⠀⠀⠀⠀⠀⠀⠀⠠⢈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠄⠁⠚⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⠛⢯⡞⣵⣋⠆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠁⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⠱⣍⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣠⡞⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡄⢀⣾⡇⠀⣾⣇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⣴⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣼⠁⣾⣿⡇⢰⣿⣿⠀⠀⣆⠀⠀⠀⠀⢰⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⢰⣿⠁⠀⠀⠀⠀⠀⠀⠀⠀⡀⠀⠀⣼⡏⢰⣿⣿⠇⣾⣿⣿⡆⠀⣿⠀⠀⠀⠀⢸⣿⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠈⡇⠀⠀⠀⠀⠀⠀⠀⠀⠰⠃⠀⠒⠛⠃⠚⠿⣿⢰⣿⣿⣿⡇⣤⣿⣤⣶⣦⣀⢼⣿⣧⠀⢰⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⢠⣶⢰⣿⣿⣿⣧⡹⢓⣾⣾⣿⣿⣿⣧⣿⣿⣿⣿⣋⣁⣀⣀⣀⣁⠘⠃⢀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣾⡟⢋⠁⡀⠀⠉⠙⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠱⣚⣭⡿⢿⣿⣷⣦⣄⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⡄⢠⣆⠀⠀⠀⠀⣿⣏⡀⣾⠀⠀⠀⠀⣰⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡏⣁⠀⢠⠀⠀⠉⠻⢿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⢠⢇⣾⣿⣷⠀⠀⠀⣿⣿⣿⣞⡓⠥⠬⣒⣷⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢿⠀⠀⠀⠀⠀⣦⠈⢳⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⢸⣾⣿⣿⣿⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣮⡢⢄⡀⠤⠾⢧⣦⣼⠇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⢟⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣶⣶⣶⣿⣿⡿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⣿⢁⣿⣿⠇⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣏⢾⡅⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠆⣼⣿⣿⣦⣾⠀⠀⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣾⣷⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀⠀⢀⠰⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⣼⣻⢿⣯⡿⣟⠇⠀⡜⢿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠇⠀⠀⠀⠀⠀⠌⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⣰⢧⡟⡿⣾⡽⢏⣿⣾⣿⡌⢻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣛⣻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡟⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠐⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⢀⡰⣣⢻⡜⣯⢳⡝⣼⣿⣿⣿⣿⣆⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡀⢂⠐⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⢠⠎⡵⢣⢧⡹⣜⢣⣿⣿⣿⣿⣿⣿⣿⣷⡌⠻⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀⠀⠀⠀⠀⠀⢀⠂⠔⡀⢂⠐⡀⢂⠠⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠡⢚⠴⣉⠦⡑⢎⢣⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⣙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠟⡩⠂⠀⠀⠀⠀⠀⣀⡔⢦⠃⢈⠐⡀⢂⠐⠠⠀⠄⠂⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠁⠎⡰⢡⠙⡌⣸⣿⣿⣿⣿⣿⣿⣿⣿⠿⠿⠟⠒⠌⠻⢿⣿⣿⣿⣿⣿⣿⣿⣿⠿⠛⠉⠀⠈⠀⠀⠀⠀⠀⣀⠶⡱⢎⢧⢋⠀⡐⢀⠂⠌⢀⠂⢀⠂⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠁⠢⠑⡨⣟⠿⠟⠟⠋⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠛⠛⠟⠛⠋⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢴⡩⢞⡱⢫⠜⡪⢅⠀⠂⠄⠂⠠⠀⠂⢀⠐⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⢢⡙⢦⡙⡔⢣⠈⢀⠂⠈⡀⠐⠀⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢤⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠂⠴⢉⠆⡁⠀⡀⠁⢀⠐⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠐⠡⠀⠀⠐⠀⠀⠀⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
  ⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠂⠈⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀

 ██╗    ██╗ █████╗ ████████╗ ██████╗██╗  ██╗ █████╗ ██████╗ ██████╗ 
 ██║    ██║██╔══██╗╚══██╔══╝██╔════╝██║  ██║██╔══██╗██╔══██╗██╔══██╗
 ██║ █╗ ██║███████║   ██║   ██║     ███████║███████║██████╔╝██████╔╝
 ██║███╗██║██╔══██║   ██║   ██║     ██╔══██║██╔══██║██╔══██╗██╔══██╗
 ╚███╔███╔╝██║  ██║   ██║   ╚██████╗██║  ██║██║  ██║██║  ██║██║  ██║
  ╚══╝╚══╝ ╚═╝  ╚═╝   ╚═╝    ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝  ╚═╝

  Layer:%s Starting now. Get ready!

  Thank you for running my (spaghetti) code on your system.`+"\n\n", version)

	// Ensure data dir exists
	err = ensureDirExists(config.DataPath)
	if err != nil {
		log.Fatal("Failed to create data dir:", err)
	}

	cfg, err := config.Get()
	if err != nil {
		log.Fatal("Failed to get server config!", err)
	}

	// Check if we want to be in DEV or PROD
	isProd := true
	if os.Getenv("MODE") == "DEV" {
		slog.Info("Starting in DEV mode")
		isProd = false
	}

	// Create our database connection.
	// Migrations are ran before our connection is returned for use.
	db, err := database.Open()
	if err != nil {
		slog.Error("Database initialization failed!", "error", err)
		os.Exit(1)
	}

	// Add a logging level hook that updates db Logger mode too, so in debug,
	// we can log ALL db queries to stdout (mainly for development, but could
	// potentially help someone debugging their instance).
	logging.AddLevelChangedHook(func(debug bool) {
		if debug {
			// Setting gorm logger to `Info` makes it log all queries, which is
			// what we want for debug mode.
			db.Logger = db.Logger.LogMode(logger.Info)
		} else {
			// When not in debug, we go back to `Warn`, which is the default
			// level for db.
			db.Logger = db.Logger.LogMode(logger.Warn)
		}
	})
	logging.Level(cfg.DEBUG)

	// Setup the db (migrations, etc). This step purposefully comes after we
	// setup logging above.
	if err := database.Setup(db); err != nil {
		slog.Error("Database setup failed!", "error", err)
		os.Exit(1)
	}

	if isProd {
		gin.SetMode(gin.ReleaseMode)
	}
	gin.DefaultWriter = multiw
	gine := gin.Default()

	// Register our custom validators
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("validsearchtype", domain.ValidSearchType)
		v.RegisterValidation("validdiscoverfilter", domain.ValidDiscoverFilter)
	}

	corsCfg := cors.Config{
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"X-CSRF-Token",
			"Authorization",
			"accept",
			"origin",
			"Cache-Control",
			"X-Requested-With",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"watcharr-lastviewedseason-saved",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	if allowedOriginsEnv := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); allowedOriginsEnv != "" {
		origins := strings.Split(allowedOriginsEnv, ",")
		for i := range origins {
			origins[i] = strings.TrimSpace(origins[i])
		}
		corsCfg.AllowOrigins = origins
	} else {
		corsCfg.AllowOriginFunc = func(origin string) bool {
			return true
		}
	}
	gine.Use(cors.New(corsCfg))

	if isProd {
		registerStaticUI(gine)
	}
	api := gine.Group("/api")
	api.Use(gzip.Gzip(gzip.DefaultCompression))

	// Health check endpoint for Docker, orchestrators and uptime monitoring
	api.GET("/health", func(c *gin.Context) {
		dbStatus := "connected"
		sqlDB, err := db.DB()
		if err != nil || sqlDB.Ping() != nil {
			dbStatus = "disconnected"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":   "degraded",
				"version":  version,
				"uptime":   time.Since(startTime).String(),
				"database": dbStatus,
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"version":  version,
			"uptime":   time.Since(startTime).String(),
			"database": dbStatus,
		})
	})

	br := router.NewBaseRouter(db, api, cfg)

	tmdbService := tmdb.NewTMDB(cfg.TMDB_KEY)
	tmdbService.DefaultLanguage = func() string {
		return cfg.GetDefaultLanguage()
	}

	contentService := content.NewService(db, tmdbService)
	tmdbService.AddContentProvider(contentService)
	plexService := plex.NewService(cfg)
	authService := auth.NewService(db, cfg, plexService)
	authTrustedHeaderService := auth.NewTrustedHeaderService(db, cfg, authService)
	userService := user.NewService(db)
	userManageService := user.NewManageService(db)
	gameService := game.NewService(db, &br.Cfg.TWITCH)
	watchedService := watched.NewService(db, contentService, gameService, userService)
	watchedSeasonService := season.NewService(
		db,
		watchedService,
		userService,
		tmdbService,
	)
	watchedEpisodeService := episode.NewService(
		db,
		watchedService,
		watchedSeasonService,
		tmdbService,
		userService,
	)
	jellyfinService := jellyfin.NewService(cfg)
	jellyfinSyncService := jellyfin.NewSyncService(
		db,
		cfg,
		jellyfinService,
		watchedService,
		watchedSeasonService,
		watchedEpisodeService,
	)
	jellyfinWebhookService := jellyfin.NewWebhookService(
		cfg,
		jellyfinService,
		authService,
		userService,
		watchedService,
		watchedEpisodeService,
	)
	plexSyncService := plex.NewSyncService(
		db,
		plexService,
		watchedService,
		watchedSeasonService,
		watchedEpisodeService,
	)
	featureService := feature.NewService(cfg)
	profileService := profile.NewService(db)
	followService := follow.NewService(db)
	tagService := tag.NewService(db, watchedService)
	searchService := search.NewService(db, br.Cfg, tmdbService, watchedService)
	discoverService := discover.NewService(db, br.Cfg, tmdbService)
	importService := imprt.NewService(
		db,
		watchedService,
		watchedSeasonService,
		watchedEpisodeService,
		tmdbService,
		tagService,
		searchService,
	)
	importTraktService := imprt.NewTraktService(importService)

	auth.NewRouter(br, authService, authTrustedHeaderService).AddRoutes()
	content.NewRouter(br, contentService, watchedService, tmdbService).AddRoutes()
	watched.NewRouter(br, watchedService).AddRoutes()
	season.NewRouter(br, watchedSeasonService).AddRoutes()
	episode.NewRouter(br, watchedEpisodeService).AddRoutes()
	activity.NewRouter(br).AddRoutes()
	profile.NewRouter(br, profileService).AddRoutes()
	jellyfin.
		NewRouter(
			br,
			jellyfinService,
			jellyfinSyncService,
			jellyfinWebhookService,
		).
		AddRoutes()
	plex.NewRouter(br, plexSyncService).AddRoutes()
	user.NewRouter(br, userService, userManageService).AddRoutes()
	follow.NewRouter(br, followService).AddRoutes()
	imprt.NewRouter(br, importService, importTraktService).AddRoutes()
	server.NewRouter(br, plexService, authTrustedHeaderService, userManageService).AddRoutes()
	feature.NewRouter(br, featureService).AddRoutes()
	arr.NewRouter(br, contentService).AddRoutes()
	seerr.NewRouter(br).AddRoutes()
	job.NewRouter(br).AddRoutes()
	task.NewRouter(br).AddRoutes()
	tag.NewRouter(br, tagService).AddRoutes()
	game.NewRouter(br, gameService, watchedService).AddRoutes()
	search.NewRouter(br, searchService, watchedService).AddRoutes()
	discover.NewRouter(br, discoverService, watchedService).AddRoutes()
	calendarService := calendar.NewService(db, cfg, tmdbService)
	calendar.NewRouter(br, calendarService).AddRoutes()
	integrationService := integration.NewService(db, cfg, tmdbService, calendarService)
	integration.NewRouter(br, integrationService).AddRoutes()
	franchiseService := franchise.NewService(db, cfg, tmdbService, watchedService)
	franchise.NewRouter(br, franchiseService).AddRoutes()
	pushService := push.NewService(db)
	push.NewRouter(br, pushService).AddRoutes()
	backupService := backup.NewService(db, cfg)
	backup.NewRouter(br, backupService).AddRoutes()
	img.NewRouter(br).AddRoutes()

	// Only add setup routes if there are no users found in db.
	var userCount int64
	if uresp := db.Model(&entity.User{}).Count(&userCount); uresp.Error == nil {
		if userCount != 0 {
			slog.Debug("registered users found.. skipped creating setup routes.")
		} else {
			slog.Info("No users found.. creating setup routes.")
			setup.NewRouter(br, authService).AddRoutes()
		}
	} else {
		slog.Error("Failed to check if any users exist.. not registering setup routes",
			"error", uresp.Error)
	}

	go taskl.SetupTasks(cfg, db, calendarService, backupService)
	go contentService.BackfillMissingMetadata()

	srv := &http.Server{
		Addr:              "0.0.0.0:3080",
		Handler:           gine.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      60 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		slog.Info("Watcharr is listening on 0.0.0.0:3080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("ListenAndServe failed!", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	slog.Info("Shutting down Watcharr gracefully...")

	// Stop recurring background tasks before closing HTTP/DB
	taskl.ShutdownTasks()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("Server forced to shutdown:", "error", err)
	}

	// Close database connections cleanly
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}

	slog.Info("Watcharr exited cleanly.")
}

// registerStaticUI serves the prebuilt static SvelteKit SPA directly from Go without Node.js.
func registerStaticUI(gine *gin.Engine) {
	uiDir := ""
	for _, candidate := range []string{"ui", "../build", "build"} {
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			if _, err := os.Stat(filepath.Join(candidate, "index.html")); err == nil {
				uiDir = candidate
				break
			}
			// Fallback if build/client exists
			clientSub := filepath.Join(candidate, "client")
			if infoSub, err := os.Stat(clientSub); err == nil && infoSub.IsDir() {
				uiDir = clientSub
				break
			}
		}
	}

	if uiDir == "" {
		slog.Warn("registerStaticUI: No static UI directory found (checked ./ui and ./build). Serving API only.")
		return
	}

	absUIDir, err := filepath.Abs(uiDir)
	if err != nil {
		absUIDir = uiDir
	}
	indexFile := filepath.Join(absUIDir, "index.html")
	slog.Info("Serving static SPA UI directly from Go", "dir", absUIDir)

	gine.NoRoute(func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasPrefix(reqPath, "/api/") || reqPath == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "API endpoint not found"})
			return
		}

		// Clean and resolve path safely inside absUIDir
		cleanRel := filepath.Clean(strings.TrimPrefix(reqPath, "/"))
		if cleanRel != "." && cleanRel != "" && !strings.HasPrefix(cleanRel, "..") {
			targetPath := filepath.Join(absUIDir, cleanRel)
			if strings.HasPrefix(targetPath, absUIDir) {
				if info, err := os.Stat(targetPath); err == nil && !info.IsDir() {
					if strings.HasPrefix(reqPath, "/_app/immutable/") {
						c.Header("Cache-Control", "public, max-age=31536000, immutable")
					} else if reqPath == "/sw.js" || strings.HasPrefix(reqPath, "/workbox-") || reqPath == "/manifest.webmanifest" || reqPath == "/_app/version.json" {
						c.Header("Cache-Control", "no-cache")
					} else {
						c.Header("Cache-Control", "public, max-age=86400")
					}
					c.File(targetPath)
					return
				}
			}
		}

		// SPA fallback to index.html for client-side routes
		c.Header("Cache-Control", "no-cache")
		c.File(indexFile)
	})
}

func ensureDirExists(dir string) error {
	err := os.MkdirAll(dir, 0764)
	if err != nil {
		return err
	}
	return nil
}
