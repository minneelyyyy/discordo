package app

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/ayn2op/discordo/internal/clipboard"
	"github.com/ayn2op/discordo/internal/config"
	"github.com/ayn2op/discordo/internal/consts"
	"github.com/ayn2op/discordo/internal/keyring"
	"github.com/ayn2op/discordo/internal/ui/accounts"
	"github.com/ayn2op/discordo/internal/ui/chat"
	"github.com/ayn2op/discordo/internal/ui/login"
	"github.com/ayn2op/tview"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/gdamore/tcell/v3"
)

type App struct {
	inner    *tview.Application
	chatView *chat.View
	cfg      *config.Config
}

func New(cfg *config.Config) *App {
	tview.Styles = tview.Theme{}
	app := &App{
		inner: tview.NewApplication(),
		cfg:   cfg,
	}

	if err := clipboard.Init(); err != nil {
		slog.Error("failed to init clipboard", "err", err)
	}

	app.inner.SetInputCapture(app.onInputCapture)
	return app
}

func (a *App) Run() error {
	screen, err := tcell.NewScreen()
	if err != nil {
		return fmt.Errorf("failed to create screen: %w", err)
	}

	if err := screen.Init(); err != nil {
		return fmt.Errorf("failed to init screen: %w", err)
	}

	if a.cfg.Mouse {
		screen.EnableMouse()
	}

	screen.SetTitle(consts.Name)
	screen.EnablePaste()
	screen.EnableFocus()
	a.inner.SetScreen(screen)

	tokenEnv := os.Getenv("DISCORDO_TOKEN")
	if tokenEnv == "" {
		accs, err := keyring.GetAccounts()
		if err != nil {
			slog.Info("failed to retrieve tokens from keyring", "err", err)
		}

		if len(accs) > 0 {
			users := make([]accounts.AcctInfo, len(accs))

			for _, acc := range accs {
				client := api.NewClient(acc.Token)
				user, err := client.Me()
				if err != nil {
					continue
				}

				users = append(users, accounts.AcctInfo{
					User:  user,
					Token: acc.Token,
				})
			}

			accountsForm := accounts.NewForm(a.inner, a.cfg, users, func(token string) {
				if err := a.showChatView(token); err != nil {
					slog.Error("failed to show chat view", "err", err)
				}
			})
			a.inner.SetRoot(accountsForm)
		} else {
			loginForm := login.NewForm(a.inner, a.cfg, func(token string) {
				if err := a.showChatView(token); err != nil {
					slog.Error("failed to show chat view", "err", err)
				}
			})
			a.inner.SetRoot(loginForm)
		}
	} else {
		tokens := strings.Split(tokenEnv, ",")

		if len(tokens) == 1 {
			if err := a.showChatView(tokens[0]); err != nil {
				return err
			}
		} else {
			users := make([]accounts.AcctInfo, len(tokens))

			for _, token := range tokens {
				client := api.NewClient(token)
				user, err := client.Me()
				if err != nil {
					continue
				}

				users = append(users, accounts.AcctInfo{
					User:  user,
					Token: token,
				})
			}

			accountsForm := accounts.NewForm(a.inner, a.cfg, users, func(token string) {
				if err := a.showChatView(token); err != nil {
					slog.Error("failed to show chat view", "err", err)
				}
			})
			a.inner.SetRoot(accountsForm)
		}
	}

	return a.inner.Run()
}

func (a *App) showChatView(token string) error {
	a.chatView = chat.NewView(a.inner, a.cfg, a.quit)
	if err := a.chatView.OpenState(token); err != nil {
		return err
	}
	a.inner.SetRoot(a.chatView)
	return nil
}

func (a *App) quit() {
	if a.chatView != nil {
		if err := a.chatView.CloseState(); err != nil {
			slog.Error("failed to close the session", "err", err)
		}
	}

	a.inner.Stop()
}

func (a *App) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case a.cfg.Keybinds.Quit:
		a.quit()
		return nil
	case "Ctrl+C":
		// https://github.com/ayn2op/tview/blob/a64fc48d7654432f71922c8b908280cdb525805c/application.go#L153
		return tcell.NewEventKey(tcell.KeyCtrlC, "", tcell.ModNone)
	}

	return event
}
