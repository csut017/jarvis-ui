package main

import (
	"context"
	"fmt"

	texttospeech "cloud.google.com/go/texttospeech/apiv1"
	texttospeechpb "google.golang.org/genproto/googleapis/cloud/texttospeech/v1"
)

func generateSpeech(text, voice string) (*texttospeechpb.SynthesizeSpeechResponse, error) {
	ctx := context.Background()
	client, err := texttospeech.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("Unable to generate speech context: %v", err)
	}

	voiceGender := texttospeechpb.SsmlVoiceGender_NEUTRAL
	switch voice {
	case "male":
		voiceGender = texttospeechpb.SsmlVoiceGender_MALE

	case "female":
		voiceGender = texttospeechpb.SsmlVoiceGender_FEMALE
	}

	req := texttospeechpb.SynthesizeSpeechRequest{
		Input: &texttospeechpb.SynthesisInput{
			InputSource: &texttospeechpb.SynthesisInput_Text{Text: text},
		},
		Voice: &texttospeechpb.VoiceSelectionParams{
			LanguageCode: "en-GB",
			SsmlGender:   voiceGender,
		},
		AudioConfig: &texttospeechpb.AudioConfig{
			AudioEncoding: texttospeechpb.AudioEncoding_MP3,
		},
	}

	resp, err := client.SynthesizeSpeech(ctx, &req)
	if err != nil {
		return nil, fmt.Errorf("Unable to synthesize speech: %v", err)
	}

	return resp, nil
}
