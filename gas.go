package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/prometheus/client_golang/prometheus"

	"github.com/rssnyder/discord-stock-ticker/utils"
)

// Gas represents the gas data
type Gas struct {
	Network   string        `json:"network"`
	Nickname  bool          `json:"nickname"`
	Frequency int           `json:"frequency"`
	APIToken  string        `json:"api_token"`
	ClientID  string        `json:"client_id"`
	Token     string        `json:"discord_bot_token"`
	close     chan struct{} `json:"-"`
}

// label returns a human readble id for this bot
func (g *Gas) label() string {
	label := strings.ToLower(g.Network)
	if len(label) > 32 {
		label = label[:32]
	}
	return label
}

// Shutdown sends a signal to shut off the goroutine
func (g *Gas) Shutdown() {
	g.close <- struct{}{}
}

// watchGasPrice gets gas prices and rotates through levels
func (g *Gas) watchGasPrice() {

	// create a new discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + g.Token)
	if err != nil {
		logger.Errorf("Error creating Discord session (%s): %s", g.ClientID, err)
		lastUpdate.With(prometheus.Labels{"type": "gas", "ticker": g.Network, "guild": "None"}).Set(0)
		return
	}

	// show as online
	err = dg.Open()
	if err != nil {
		logger.Errorf("Error opening discord connection (%s): %s", g.ClientID, err)
		lastUpdate.With(prometheus.Labels{"type": "gas", "ticker": g.Network, "guild": "None"}).Set(0)
		return
	}

	// Get guides for bot
	guilds, err := dg.UserGuilds(100, "", "")
	if err != nil {
		logger.Errorf("Error getting guilds: %s", err)
		g.Nickname = false
	}
	if len(guilds) == 0 {
		g.Nickname = false
	}

	// check for frequency override
	// set to ten min to avoid lockout
	if *frequency != 0 {
		g.Frequency = 600
	}

	// perform management operations
	if *managed {
		setName(dg, g.label())
	}

	logger.Infof("Watching gas price for %s", g.Network)
	ticker := time.NewTicker(time.Duration(g.Frequency) * time.Second)

	g.close = make(chan struct{}, 1)

	// watch gas price
	for {

		select {
		case <-g.close:
			logger.Infof("Shutting down price watching for %s", g.Network)
			return
		case <-ticker.C:
			// get gas prices
			gasPrices, err := utils.GetGasPrices(g.Network, g.APIToken)
			if err != nil {
				logger.Errorf("Error getting rates: %s", err)
				continue
			}

			nickname := fmt.Sprintf("⚡ %d 🤔 %d 🐌 %d", gasPrices.Instant, gasPrices.Fast, gasPrices.Standard)

			// change nickname
			if g.Nickname {

				for _, gu := range guilds {

					err = dg.GuildMemberNickname(gu.ID, "@me", nickname)
					if err != nil {
						logger.Errorf("Error updating nickname: %s", err)
						continue
					} else {
						logger.Debugf("Set nickname in %s: %s", gu.Name, nickname)
					}
					lastUpdate.With(prometheus.Labels{"type": "gas", "ticker": g.Network, "guild": gu.Name}).SetToCurrentTime()
					time.Sleep(time.Duration(g.Frequency) * time.Second)
				}

				err = dg.UpdateWatchStatus(0, "Fast, Avg, Slow")
				if err != nil {
					logger.Errorf("Unable to set activity: %s", err)
				} else {
					logger.Debugf("Set activity")
				}
			} else {

				err = dg.UpdateWatchStatus(0, nickname)
				if err != nil {
					logger.Errorf("Unable to set activity: %s", err)
				} else {
					logger.Debugf("Set activity: %s", nickname)
					lastUpdate.With(prometheus.Labels{"type": "gas", "ticker": g.Network, "guild": "None"}).SetToCurrentTime()
				}
			}
		}
	}
}
