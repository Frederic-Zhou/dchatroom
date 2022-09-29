package view

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var mainFlex *tview.Flex
var inputText *tview.InputField
var messageView *tview.TextView
var infoView *tview.TextView
var commandList *tview.DropDown
var app *tview.Application
var recipientsList *tview.DropDown

var commandValues, commandNames []string

var enterFunc func(text string)

func init() {

	commandValues = []string{"", "/sub ", "/aka ", "/accept ", "/reject "}
	commandNames = []string{"Back to input", "/sub: Sub a topic", "/aka: Set AKA", "/accept: accept recipient", "/reject: reject recipient"}

	app = tview.NewApplication()
	mainFlex = tview.NewFlex()
	inputText = tview.NewInputField()
	messageView = tview.NewTextView()
	infoView = tview.NewTextView()
	commandList = tview.NewDropDown()
	recipientsList = tview.NewDropDown()

	messageView.
		SetDynamicColors(true).
		SetBorder(false).
		SetBackgroundColor(tcell.ColorBlack)

	infoView.
		SetDynamicColors(true).
		SetBorder(false).
		SetBackgroundColor(tcell.ColorGreenYellow)

	inputText.
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetBackgroundColor(tcell.ColorAntiqueWhite)

	commandList.
		SetOptions(commandNames, nil).
		SetBackgroundColor(tcell.ColorBlanchedAlmond)

	recipientsList.
		SetOptions([]string{"Empty"}, nil).
		SetBackgroundColor(tcell.ColorBlanchedAlmond)

	mainFlex.SetDirection(tview.FlexRow).
		AddItem(infoView, 1, 0, false).
		AddItem(messageView, 0, 1, false).
		AddItem(inputText, 1, 0, true)

	setEvents()

}

func setEvents() {
	inputText.SetChangedFunc(func(text string) {
		if text == "/" {
			mainFlex.AddItem(commandList, 1, 0, false)
			app.SetFocus(commandList)
		}
		if text == "/reject " || text == "/accept " {
			mainFlex.AddItem(recipientsList, 1, 0, false)
			app.SetFocus(recipientsList)
		}
	})

	inputText.SetDoneFunc(func(key tcell.Key) {
		switch key {
		case tcell.KeyEnter:
			if enterFunc != nil {
				enterFunc(inputText.GetText())
			}

			inputText.SetText("")
		case tcell.KeyEscape:
			inputText.SetText("")
		}

	})

	messageView.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(inputText)
	})

	infoView.SetDoneFunc(func(key tcell.Key) {
		app.SetFocus(inputText)
	})

	commandList.SetSelectedFunc(func(text string, index int) {
		app.SetFocus(inputText)
		mainFlex.RemoveItem(commandList)
		inputText.SetText(commandValues[index])
	})

	recipientsList.SetSelectedFunc(func(text string, index int) {
		app.SetFocus(inputText)
		mainFlex.RemoveItem(recipientsList)
		inputText.SetText(text)
	})
}

func Run(ef func(text string)) {
	enterFunc = ef
	if err := app.SetRoot(mainFlex, true).EnableMouse(true).SetFocus(inputText).Run(); err != nil {
		panic(err)
	}
}

func AddMessage(text []byte) {

	go func(txt []byte) {
		app.QueueUpdateDraw(func() {
			messageView.Write(append(txt, '\n'))
			messageView.ScrollToEnd()
		})
	}(text)

}

func SetInfoView(info string) {
	go func(inf string) {
		app.QueueUpdateDraw(func() {
			infoView.SetText(inf)
		})
	}(info)
}

func SetRecipientListOptions(recipients []string) {

	go func(recs []string) {
		app.QueueUpdateDraw(func() {

			recs = append([]string{"Back to input"}, recs...)

			recipientsList.SetOptions(recs, nil).
				SetSelectedFunc(func(text string, index int) {
					app.SetFocus(inputText)
					mainFlex.RemoveItem(recipientsList)

					if index == 0 {
						inputText.SetText("")
						return
					}

					inputText.SetText(inputText.GetText() + strings.Split(text, ":")[0])
				})
		})
	}(recipients)
}
