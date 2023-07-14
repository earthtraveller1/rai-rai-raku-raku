package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/creack/pty"
)

var terminal *os.File
var terminalReady bool = false

var serverName string

func main() {
    if len( os.Args ) < 2 {
        fmt.Fprintf(os.Stderr, "[USAGE]: %s <serverName>", os.Args[0])
        return
    }

    serverName = os.Args[1]

    token, error := ioutil.ReadFile("user/token.txt")
    if error != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]: Failed to read the token file. Maybe it doesn't exist?")
        return
    }

    discord, error := discordgo.New("Bot " + strings.TrimSpace(string(token)))
    if error != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]: Failed to login to Discord. Error:", error)
        return
    }

    discord.AddHandler(onReady)
    discord.AddHandler(onMessageCreate)
    discord.Identify.Intents = discordgo.IntentGuildMessages

    error = discord.Open()
    if error != nil {
        fmt.Fprintln(os.Stderr, "[ERROR]: Failed to open a connection to Discord. Error", error)
        return
    }

    defer discord.Close()

    signalChannel := make(chan os.Signal, 1)
    signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
    <-signalChannel

    terminal.Close()
}

func onReady(pSession *discordgo.Session, pReadyEvent *discordgo.Ready) {
    mainChannel := "1076539147327643689"

    command := exec.Command("./server.sh", serverName)

    var error error
    terminal, error = pty.Start(command)
    if error != nil {
        fmt.Fprintln(os.Stderr, "failed to open fake terminal bc of ", error)
        return
    }

    terminalReady = true

    scanner := bufio.NewScanner(terminal)

    for scanner.Scan() {
        pSession.ChannelMessageSend(mainChannel, "`" + scanner.Text() + "`")
    }

    fmt.Println("[INFO]: Okay, we are done now. You can close this terminal now.")
}

func onMessageCreate(pSession *discordgo.Session, pMessageCreateEvent *discordgo.MessageCreate) {
    mainChannel := "1076539147327643689"

    messageFromMainChannel := pMessageCreateEvent.ChannelID == mainChannel 
    notFromMyself := pSession.State.User.ID != pMessageCreateEvent.Author.ID 
    fromHello56721 := pMessageCreateEvent.Author.ID == "650439182204010496"

    if terminalReady && messageFromMainChannel && notFromMyself && fromHello56721 {
        command := (pMessageCreateEvent.Content)
        _, error := terminal.WriteString(command + "\n")

        if error != nil {
            fmt.Fprintln(os.Stderr, "[ERROR]: ", error.Error())
            return
        }
    }
}