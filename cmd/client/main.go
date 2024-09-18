package main

import (
	"context"
	"embed"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"sync"
	"time"

	tictactoev1 "TicTacToe/api/tictactoe"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	client        tictactoev1.GameServiceClient
	gameData      *tictactoev1.GameData
	gameID        string
	playerName    string
	playerID      string
	playerSymbol  string
	mu            sync.Mutex
	moveSound     *beep.Buffer
	buttonSound   *beep.Buffer
	confettiImage fyne.Resource
	xImage        fyne.Resource
	oImage        fyne.Resource
	//go:embed resources/*
	resources embed.FS
)

var localGameState struct {
	board [9]string
}

func main() {
	myApp := app.New()
	myApp.Settings().SetTheme(&myTheme{})
	myWindow := myApp.NewWindow("TicTacToe")

	sampleRate := beep.SampleRate(44100)
	err := speaker.Init(sampleRate, sampleRate.N(time.Second/10))
	if err != nil {
		panic("cannot load speaker")
	}

	// Load resources
	moveSound = loadSound("move.wav")
	buttonSound = loadSound("button.wav")
	xImage = loadResourceFromPath("x.png")
	oImage = loadResourceFromPath("o.png")

	// Create the start screen
	startScreen := createStartScreen(myWindow)
	myWindow.SetContent(startScreen)

	myWindow.Resize(fyne.NewSize(400, 400))
	myWindow.CenterOnScreen()
	myWindow.ShowAndRun()
}

// Custom theme for the application
type myTheme struct{}

func (m myTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // Dark background
	case theme.ColorNameButton:
		return color.NRGBA{R: 38, G: 38, B: 38, A: 255} // Dark buttons
	case theme.ColorNameDisabledButton:
		return color.NRGBA{R: 58, G: 58, B: 58, A: 255} // Disabled buttons
	case theme.ColorNameHover:
		return color.NRGBA{R: 58, G: 58, B: 58, A: 255} // Hover effect
	case theme.ColorNameForeground:
		return color.NRGBA{R: 230, G: 230, B: 230, A: 255} // Light text
	case theme.ColorNameInputBackground:
		return color.NRGBA{R: 28, G: 28, B: 28, A: 255} // Input background
	case theme.ColorNameInputBorder:
		return color.NRGBA{R: 58, G: 58, B: 58, A: 255} // Input border
	case theme.ColorNamePlaceHolder:
		return color.NRGBA{R: 128, G: 128, B: 128, A: 255} // Placeholder text
	case theme.ColorNameFocus:
		return color.NRGBA{R: 70, G: 130, B: 180, A: 255} // Steel blue focus
	case theme.ColorNameOverlayBackground:
		return color.NRGBA{R: 18, G: 18, B: 18, A: 255} // Dialog
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m myTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m myTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m myTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

func (m myTheme) Variant() fyne.ThemeVariant {
	return 1
}

// Start screen where the player logs in
func createStartScreen(window fyne.Window) fyne.CanvasObject {
	title := widget.NewLabelWithStyle("Welcome to TicTacToe!", fyne.TextAlignCenter, fyne.TextStyle{Bold: true, Italic: true})

	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("Enter your name")

	loginButton := widget.NewButton("Login", func() {
		playSound(buttonSound)
		if nameEntry.Text != "" {
			playerName = nameEntry.Text
			connectToServer()
			err := loginPlayer()
			if err != nil {
				dialog.ShowError(err, window)
				return
			}
			showGameOptionsScreen(window)
		}
	})

	content := container.NewVBox(
		title,
		nameEntry,
		loginButton,
	)
	return container.NewCenter(content)
}

// Screen where the player chooses to create or join a game
func showGameOptionsScreen(window fyne.Window) {
	title := widget.NewLabelWithStyle(fmt.Sprintf("Welcome, %s!", playerName), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	//title.TextSize = 24

	createGameButton := widget.NewButton("Create Game", func() {
		playSound(buttonSound)
		showCreateGameScreen(window)
	})

	joinGameButton := widget.NewButton("Join Game", func() {
		playSound(buttonSound)
		showJoinGameScreen(window)
	})

	backButton := widget.NewButton("Back", func() {
		playSound(buttonSound)
		window.SetContent(createStartScreen(window))
	})

	content := container.NewVBox(
		title,
		createGameButton,
		joinGameButton,
		backButton,
	)
	window.SetContent(container.NewCenter(content))
}

// Screen to create a new game with a password
func showCreateGameScreen(window fyne.Window) {
	title := widget.NewLabelWithStyle("Create a New Game", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	//title.TextSize = 24

	passwordEntry := widget.NewEntry()
	passwordEntry.SetPlaceHolder("Enter password")

	errorLabel := widget.NewLabel("")
	errorLabel.Alignment = fyne.TextAlignCenter
	errorLabel.Hide()

	createButton := widget.NewButton("Create", func() {
		playSound(buttonSound)
		if len(passwordEntry.Text) < 4 {
			errorLabel.SetText("Password must be at least 4 characters long")
			errorLabel.Show()
		} else {
			errorLabel.Hide()
			err := createGame(passwordEntry.Text)
			if err != nil {
				errorLabel.SetText(err.Error())
				errorLabel.Show()
				return
			}
			showGameBoard(window)
		}
	})

	backButton := widget.NewButton("Back", func() {
		playSound(buttonSound)
		showGameOptionsScreen(window)
	})

	content := container.NewVBox(
		title,
		passwordEntry,
		errorLabel,
		createButton,
		backButton,
	)
	window.SetContent(container.NewCenter(content))
}

// Screen to join an existing game
func showJoinGameScreen(window fyne.Window) {
	title := widget.NewLabelWithStyle("Join an Existing Game", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	//title.TextSize = 24

	gameIDEntry := widget.NewEntry()
	gameIDEntry.SetPlaceHolder("Enter game ID")

	passwordEntry := widget.NewEntry()
	passwordEntry.SetPlaceHolder("Enter game password")

	errorLabel := widget.NewLabel("")
	errorLabel.Alignment = fyne.TextAlignCenter
	errorLabel.Hide()

	joinButton := widget.NewButton("Join", func() {
		playSound(buttonSound)
		if gameIDEntry.Text == "" || passwordEntry.Text == "" {
			errorLabel.SetText("Game ID and Password cannot be empty")
			errorLabel.Show()
		} else {
			errorLabel.Hide()
			err := joinGame(gameIDEntry.Text, passwordEntry.Text)
			if err != nil {
				errorLabel.SetText(fmt.Sprintf("Failed to join game: %v", err))
				errorLabel.Show()
				return
			}
			showGameBoard(window)
		}
	})

	backButton := widget.NewButton("Back", func() {
		playSound(buttonSound)
		showGameOptionsScreen(window)
	})

	content := container.NewVBox(
		title,
		gameIDEntry,
		passwordEntry,
		errorLabel,
		joinButton,
		backButton,
	)
	window.SetContent(container.NewCenter(content))
}

// Main game board where the game is played
func showGameBoard(window fyne.Window) {
	boardButtons := make([]*widget.Button, 9)
	for i := 0; i < 9; i++ {
		index := i
		boardButtons[i] = widget.NewButton("", func() {
			makeMove(index)
		})
		boardButtons[i].Importance = widget.HighImportance
		boardButtons[i].Alignment = widget.ButtonAlignCenter
		//boardButtons[i].IconPlacement = widget.
		//boardButtons[i].SetMinSize(fyne.NewSize(100, 100))
	}

	boardObjects := make([]fyne.CanvasObject, len(boardButtons))
	for i, b := range boardButtons {
		boardObjects[i] = b
	}

	board := container.NewGridWithColumns(3, boardObjects...)
	paddedBoard := container.NewPadded(board)

	statusLabel := widget.NewLabel("Waiting for game to start...")
	statusLabel.Alignment = fyne.TextAlignCenter
	statusLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: true}

	currentPlayerLabel := widget.NewLabel("")
	currentPlayerLabel.Alignment = fyne.TextAlignCenter
	currentPlayerLabel.TextStyle = fyne.TextStyle{Bold: true}

	copyIDButton := widget.NewButtonWithIcon("Copy Game ID", theme.ContentCopyIcon(), func() {
		playSound(buttonSound)
		window.Clipboard().SetContent(gameID)
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Copied",
			Content: "Game ID copied to clipboard",
		})
	})

	copyPasswordButton := widget.NewButtonWithIcon("Copy Password", theme.ContentCopyIcon(), func() {
		playSound(buttonSound)
		window.Clipboard().SetContent(gameData.Password)
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Copied",
			Content: "Password copied to clipboard",
		})
	})

	buttonContainer := container.NewHBox(copyIDButton, copyPasswordButton)

	leaveButton := widget.NewButton("Leave Game", func() {
		playSound(buttonSound)
		leaveGame()
		showGameOptionsScreen(window)
	})
	leaveButton.Importance = widget.DangerImportance

	playerInfo := widget.NewLabelWithStyle(fmt.Sprintf("Player: %s (%s)", playerName, playerSymbol), fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	content := container.NewVBox(
		playerInfo,
		buttonContainer,
		paddedBoard,
		statusLabel,
		currentPlayerLabel,
		container.NewCenter(leaveButton),
	)

	window.SetContent(container.NewCenter(content))

	go listenForUpdates(func() {
		updateGameBoard(boardButtons, statusLabel, currentPlayerLabel, window)
	})
}

// Load image resources
func loadResourceFromPath(path string) fyne.Resource {
	resource, err := fyne.LoadResourceFromPath(path)
	if err != nil {
		log.Fatalf("Failed to load resource %s: %v", path, err)
	}
	return resource
}

// Load sound files into a beep.Buffer
func loadSound(path string) *beep.Buffer {
	f, err := os.Open(path)
	if err != nil {
		log.Fatalf("Failed to open sound file %s: %v", path, err)
	}
	defer f.Close()

	streamer, format, err := wav.Decode(f)
	if err != nil {
		log.Fatalf("Failed to decode sound file %s: %v", path, err)
	}
	defer streamer.Close()

	buffer := beep.NewBuffer(format)
	buffer.Append(streamer)

	return buffer
}

// Play sound from a beep.Buffer
func playSound(buffer *beep.Buffer) {
	streamer := buffer.Streamer(0, buffer.Len())
	speaker.Play(streamer)
}

func updateCell(button *widget.Button, cell string) {
	switch cell {
	case "X":
		button.SetIcon(fyne.NewStaticResource("X", xImage.Content()))
	case "O":
		button.SetIcon(fyne.NewStaticResource("O", oImage.Content()))
	}
	button.Refresh()
}

// Update the game board UI
func updateGameBoard(boardButtons []*widget.Button, statusLabel, currentPlayerLabel *widget.Label, window fyne.Window) {
	mu.Lock()
	defer mu.Unlock()

	if gameData == nil {
		return
	}

	isPlayerTurn := gameData.CurrentPlayer != nil && gameData.CurrentPlayer.PlayerId == playerID
	isGameStarted := gameData.Status == tictactoev1.GameStatus_IN_PROGRESS

	if gameData.Status == tictactoev1.GameStatus_WAITING_FOR_PLAYER {
		isGameStarted = false
	}

	for i, cell := range gameData.Board {
		if cell != localGameState.board[i] {
			updateCell(boardButtons[i], cell)
			localGameState.board[i] = cell
		}
		if cell != "" || !isPlayerTurn || !isGameStarted {
			boardButtons[i].Disable()
		} else {
			boardButtons[i].Enable()
		}

	}

	if !isGameStarted {
		statusLabel.SetText("Waiting for second player to join...")
	} else {
		statusText := fmt.Sprintf("Current player: %s", gameData.CurrentPlayer.PlayerName)
		statusLabel.SetText(statusText)
	}

	if isPlayerTurn && isGameStarted {
		currentPlayerLabel.SetText("Your turn!")
	} else {
		if isGameStarted {
			currentPlayerLabel.SetText(fmt.Sprintf("%s's turn", gameData.CurrentPlayer.PlayerName))
		} else {
			currentPlayerLabel.SetText("Waiting for opponent...")
		}
	}

	// Handle game over state
	if gameData.Status == tictactoev1.GameStatus_FINISHED {
		if gameData.Winner != "" {
			if gameData.Winner == playerName {
				showConfetti(window)
			}
			showGameOverDialog(window, fmt.Sprintf("Game over! Winner: %s", gameData.Winner))
		} else {
			showGameOverDialog(window, "Game over! It's a draw!")
		}
	}
}

// Show a dialog when the game is over
func showGameOverDialog(window fyne.Window, message string) {
	title := canvas.NewText("Game Over", color.White)
	title.TextStyle = fyne.TextStyle{Bold: true}
	title.Alignment = fyne.TextAlignCenter

	msg := canvas.NewText(message, color.White)
	msg.Alignment = fyne.TextAlignCenter

	okButton := widget.NewButton("OK", func() {
		window.Canvas().Overlays().Top().Hide()
	})

	content := container.NewVBox(
		title,
		msg,
		okButton,
	)
	modal := widget.NewModalPopUp(content, window.Canvas())
	modal.Show()
}

// Show confetti effect
func showConfetti(window fyne.Window) {
	confetti := canvas.NewImageFromResource(confettiImage)
	confetti.FillMode = canvas.ImageFillContain

	overlay := container.NewCenter(confetti)
	window.Canvas().Overlays().Add(overlay)

	// Remove the confetti after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		window.Canvas().Overlays().Remove(overlay)
	}()
}

// Connect to the game server
func connectToServer() {
	conn, err := grpc.Dial("localhost:44044", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	client = tictactoev1.NewGameServiceClient(conn)
}

// Login the player and store the player ID
func loginPlayer() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	resp, err := client.Login(ctx, &tictactoev1.LoginRequest{
		PlayerName: playerName,
	})
	if err != nil {
		return fmt.Errorf("Login failed: %v", extractErrorMessage(err))
	}
	playerID = resp.PlayerId
	return nil
}

// Create a new game on the server
func createGame(password string) error {
	ctx := contextWithPlayerID()
	resp, err := client.CreateGame(ctx, &tictactoev1.CreateGameRequest{
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("%v", extractErrorMessage(err))
	}
	gameID = resp.Id
	gameData = resp

	// Assign symbols
	playerSymbol = "X"
	return nil
}

// Join an existing game on the server
func joinGame(gameIDParam, password string) error {
	ctx := contextWithPlayerID()
	resp, err := client.JoinGame(ctx, &tictactoev1.JoinGameRequest{
		GameId:   gameIDParam,
		Password: password,
	})
	if err != nil {
		return fmt.Errorf("%v", extractErrorMessage(err))
	}
	gameData = resp
	gameID = gameIDParam

	// Assign symbols
	playerSymbol = "O"
	return nil
}

// Leave the game
func leaveGame() {
	ctx := contextWithPlayerID()
	_, err := client.LeaveGame(ctx, &tictactoev1.LeaveGameRequest{
		GameId: gameID,
	})
	if err != nil {
		log.Printf("Failed to leave game: %v", extractErrorMessage(err))
	}
	gameID = ""
	gameData = nil
	playerSymbol = ""
}

// Create a context with player-id metadata
func contextWithPlayerID() context.Context {
	md := metadata.Pairs("player-id", playerID)
	return metadata.NewOutgoingContext(context.Background(), md)
}

func extractErrorMessage(err error) string {
	if grpcErr, ok := status.FromError(err); ok {
		return grpcErr.Message()
	}
	return err.Error()
}

// Make a move on the game board
func makeMove(position int) {
	ctx := contextWithPlayerID()
	_, err := client.MakeMove(ctx, &tictactoev1.MoveRequest{
		GameId:   gameID,
		Position: int32(position),
	})
	if err != nil {
		fyne.CurrentApp().SendNotification(&fyne.Notification{
			Title:   "Move Failed",
			Content: fmt.Sprintf("%v", extractErrorMessage(err)),
		})
	} else {
		playSound(moveSound)
	}
}

// Listen for updates from the server
func listenForUpdates(updateUI func()) {
	ctx := contextWithPlayerID()
	stream, err := client.GetGameState(ctx, &tictactoev1.GameRequest{GameId: gameID})
	if err != nil {
		log.Fatalf("Failed to get game state: %v", err)
	}

	for {
		update, err := stream.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("Failed to receive update: %v", err)
			fyne.CurrentApp().SendNotification(&fyne.Notification{
				Title:   "Connection Lost",
				Content: "Attempting to reconnect...",
			})
			time.Sleep(time.Second * 2)
			// Attempt to reconnect
			go listenForUpdates(updateUI)
			return
		}
		mu.Lock()
		gameData = update
		mu.Unlock()
		updateUI()
	}
}
