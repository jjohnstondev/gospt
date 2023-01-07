package runner

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

func Run(client *spotify.Client, args []string) error {
	if len(args) == 0 {
		user, err := client.CurrentUser(context.Background())
		if err != nil {
			return fmt.Errorf("Failed to get current user")
		}
		fmt.Println("The following commands are currently supported:\nplay pause next shuffle\nhave fun", user.DisplayName)
		return nil
	}
	ctx := context.Background()
	switch args[0] {
	case "play":
		err := Play(ctx, client, args)
		if err != nil {
			return err
		}
	case "pause":
		err := Pause(ctx, client, args)
		if err != nil {
			return err
		}
	case "next":
		err := Skip(ctx, client, args)
		if err != nil {
			return err
		}
	case "shuffle":
		err := Shuffle(ctx, client, args)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Unsupported Command")
	}
	return nil
}

func Play(ctx context.Context, client *spotify.Client, args []string) error {
	err := client.Play(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Playing!")
	return nil
}

func Pause(ctx context.Context, client *spotify.Client, args []string) error {
	err := client.Pause(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Pausing!")
	return nil
}

func Skip(ctx context.Context, client *spotify.Client, args []string) error {
	err := client.Next(ctx)
	if err != nil {
		return err
	}
	fmt.Println("Skipping!")
	return nil
}

func Shuffle(ctx context.Context, client *spotify.Client, args []string) error {
	state, err := client.PlayerState(ctx)
	if err != nil {
		return fmt.Errorf("Failed to get current playstate")
	}
	err = client.Shuffle(ctx, !state.ShuffleState)
	if err != nil {
		return err
	}
	fmt.Println("Shuffle set to", !state.ShuffleState)
	return nil
}
