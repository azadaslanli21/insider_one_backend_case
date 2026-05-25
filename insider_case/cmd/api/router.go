package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insider/football-simulator/internal/domain"
	"github.com/insider/football-simulator/internal/service"
)

type apiHandler struct {
	league *service.LeagueService
}

func setupRouter(league *service.LeagueService) *gin.Engine {
	r := gin.Default()
	h := &apiHandler{league: league}

	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	r.POST("/league/init", h.initLeague)
	r.POST("/league/simulate-next", h.simulateNext)
	r.POST("/league/simulate-all", h.simulateAll)
	r.GET("/league/table", h.getTable)
	r.GET("/league/fixtures", h.getFixtures)
	r.POST("/league/reset", h.resetLeague)
	r.PATCH("/league/match/:id", h.patchMatch)
	return r
}

func (h *apiHandler) initLeague(c *gin.Context) {
	if err := h.league.InitLeague(c.Request.Context()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "initialized"})
}

func (h *apiHandler) simulateNext(c *gin.Context) {
	resp, err := h.league.SimulateNextWeek(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"results":       resp.Results,
		"standings":     resp.Standings,
		"predictions":   resp.Predictions,
		"week_id":       resp.WeekID,
		"ai_commentary": resp.AICommentary,
	})
}

func (h *apiHandler) simulateAll(c *gin.Context) {
	resp, err := h.league.SimulateAll(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *apiHandler) getTable(c *gin.Context) {
	table, err := h.league.GetTable(c.Request.Context())
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"standings": table})
}

func (h *apiHandler) getFixtures(c *gin.Context) {
	mode := c.Query("mode")
	resp, err := h.league.GetFixtures(c.Request.Context(), mode)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (h *apiHandler) resetLeague(c *gin.Context) {
	if err := h.league.ResetLeague(c.Request.Context()); err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "reset"})
}

type patchMatchRequest struct {
	HomeGoals int    `json:"home_goals"`
	AwayGoals int    `json:"away_goals"`
	Reasoning string `json:"reasoning"`
}

func (h *apiHandler) patchMatch(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_MATCH_ID",
			"message": "match id must be an unsigned integer",
		})
		return
	}

	var req patchMatchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_REQUEST_BODY",
			"message": err.Error(),
		})
		return
	}

	standings, err := h.league.OverrideMatchResult(c.Request.Context(), uint(id), req.HomeGoals, req.AwayGoals, req.Reasoning)
	if err != nil {
		writeError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"standings": standings})
}

func writeError(c *gin.Context, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		status := http.StatusInternalServerError
		switch appErr.Code {
		case "NO_MATCHES_REMAINING":
			status = http.StatusConflict
		case "SEASON_ALREADY_COMPLETE":
			status = http.StatusConflict
		case "LEAGUE_NOT_INITIALIZED":
			status = http.StatusBadRequest
		case "MATCH_NOT_FOUND", "NOT_FOUND":
			status = http.StatusNotFound
		case "INVALID_GOAL_COUNT":
			status = http.StatusBadRequest
		case "INVALID_MODE":
			status = http.StatusBadRequest
		case "DB_CONNECTION_FAILURE":
			status = http.StatusServiceUnavailable
		case "DB_ERROR":
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"code": appErr.Code, "message": appErr.Message})
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{
		"code":    "INTERNAL_ERROR",
		"message": err.Error(),
	})
}
