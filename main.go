package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	arg "github.com/alexflint/go-arg"
	shellwords "github.com/mattn/go-shellwords"
	"github.com/sirupsen/logrus"
	"google.golang.org/api/option"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

type args struct {
	LanguageCode string  `arg:"-l, --language" default:"en-US"`
	VoiceName    string  `arg:"--voice" default:"en-US-Standard-A"`
	SpeakingRate float64 `arg:"-s, --speed" default:"1.0"`
	Pitch        float64 `arg:"-p, --pitch" default:"0.0"`
	VolumeGainDB float64 `arg:"-g, --gain" default:"0.0"`
	Player       string  `arg:"--player" default:"mpv"`

	CredentialsPath string `arg:"-c, --credentials, env:GOOGLE_APPLICATION_CREDENTIALS, required" help:"application credentials file path"`
}

func main() {
	logger := logrus.StandardLogger()

	var args args
	arg.MustParse(&args)

	playerParsed, err := shellwords.Parse(args.Player)

	if err != nil {
		logger.WithError(err).Fatalf("invalid player command '%s'", args.Player)
	}

	playerParsed = append(playerParsed, "-")

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()
		synthesized, err := tts(args, input)

		if err != nil {
			logger.WithError(err).Error("failed to synthesize speech")
			continue
		}

		command := exec.Command(playerParsed[0], playerParsed[1:]...)

		stdin, err := command.StdinPipe()
		if err != nil {
			logger.WithError(err).Errorf("failed to get stdin of player command '%s'", args.Player)
		}

		go func() {
			defer stdin.Close()

			if _, err := stdin.Write(synthesized); err != nil {
				logger.WithError(err).Errorf("failed to write to stdin of command '%s'", args.Player)
			}
		}()

		if err := command.Run(); err != nil {
			logger.WithError(err).Errorf("failed to executo player command '%s'", args.Player)
		}
	}
}

func tts(args args, text string) ([]byte, error) {
	ctx := context.Background()

	client, err := texttospeech.NewClient(ctx, option.WithCredentialsFile(args.CredentialsPath))

	if err != nil {
		return nil, fmt.Errorf("failed to create text-to-speech client: %w", err)
	}

	response, err := client.SynthesizeSpeech(ctx, &texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{
				Text: text,
			},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: args.LanguageCode,
			Name:         args.VoiceName,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_OGG_OPUS,
			SpeakingRate:  args.SpeakingRate,
			Pitch:         args.Pitch,
			VolumeGainDb:  args.VolumeGainDB,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	return response.AudioContent, nil
}
