package accounts

import (
	"log/slog"

	"github.com/ayn2op/tview/layers"
	"github.com/diamondburned/arikawa/v3/api"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/gdamore/tcell/v3"

	"github.com/ayn2op/discordo/internal/config"
	"github.com/ayn2op/discordo/internal/keyring"
	"github.com/ayn2op/discordo/internal/ui"
	"github.com/ayn2op/tview"
	"golang.design/x/clipboard"
)

const (
	formLayerName  = "form"
	errorLayerName = "error"
)

type DoneFn = func(token string)

type Form struct {
	*layers.Layers
	app  *tview.Application
	cfg  *config.Config
	form *tview.Form
	done DoneFn
}

type AcctInfo struct {
	User  *discord.User
	Token string
}

func NewForm(app *tview.Application, cfg *config.Config, accts []AcctInfo, done DoneFn) *Form {
	f := &Form{
		Layers: layers.New(),
		app:    app,
		cfg:    cfg,
		form:   tview.NewForm(),
		done:   done,
	}

	for _, acct := range accts {
		token := acct.Token

		client := api.NewClient(token)
		user, err := client.Me()
		if err != nil {
			continue
		}

		f.form.AddButton(user.Username, func() {
			account := keyring.AccountInfo{
				Id:    user.ID,
				Token: token,
			}

			f.login(account)
		})
	}

	f.form.
		AddButton("New Account", f.addAccount)

	f.SetBackgroundLayerStyle(f.cfg.Theme.Dialog.BackgroundStyle.Style)
	f.AddLayer(f.form, layers.WithName(formLayerName), layers.WithResize(true), layers.WithVisible(true))
	return f
}

func (f *Form) login(account keyring.AccountInfo) {
	if f.done != nil {
		f.done(account.Token)
	}
}

func (f *Form) addAccount() {
}

func (f *Form) onError(err error) {
	slog.Error("failed to login", "err", err)

	message := err.Error()
	modal := tview.NewModal().
		SetText(message).
		AddButtons([]string{"Copy", "Close"}).
		SetDoneFunc(func(buttonIndex int, _ string) {
			if buttonIndex == 0 {
				go clipboard.Write(clipboard.FmtText, []byte(message))
			} else {
				f.RemoveLayer(errorLayerName)
			}
		})
	{
		bg := f.cfg.Theme.Dialog.Style.GetBackground()
		buttonStyle := f.cfg.Theme.Dialog.Style.Style
		if bg != tcell.ColorDefault {
			modal.SetBackgroundColor(bg)
			buttonStyle = buttonStyle.Background(bg)
		}
		fg := f.cfg.Theme.Dialog.Style.GetForeground()
		if fg != tcell.ColorDefault {
			modal.SetTextColor(fg)
			buttonStyle = buttonStyle.Foreground(fg)
		}
		// Keep button styles aligned with dialog content without hiding text.
		modal.SetButtonStyle(buttonStyle)
		modal.SetButtonActivatedStyle(buttonStyle)
	}
	f.
		AddLayer(
			ui.Centered(modal, 0, 0),
			layers.WithName(errorLayerName),
			layers.WithResize(true),
			layers.WithVisible(true),
			layers.WithOverlay(),
		).
		SendToFront(errorLayerName)
}
