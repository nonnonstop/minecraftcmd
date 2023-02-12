package main

import (
	"errors"
	"os/exec"
	"time"

	"github.com/bwmarrin/discordgo"
	"go.uber.org/zap"
)

func runCommand(command string) error {
	cmd := exec.Command("sh", "-c", command)
	err := cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	return err
}

func waitServer(config *AppConfig, expectStopping bool) error {
	success := false
	for i := 0; i < 10; i++ {
		time.Sleep(1 * time.Second)
		err := runCommand(config.CheckCmd)
		if (err != nil) == expectStopping {
			success = true
			break
		}
	}
	if !success {
		return errors.New("Failed to wait server")
	}
	return nil
}

func startServer(config *AppConfig, logger *zap.SugaredLogger) error {
	if err := runCommand(config.StartCmd); err != nil {
		return err
	}
	if err := waitServer(config, false); err != nil {
		return err
	}
	return nil
}

func stopServer(config *AppConfig, logger *zap.SugaredLogger) error {
	if err := runCommand(config.StopCmd); err != nil {
		return err
	}
	if err := waitServer(config, true); err != nil {
		return err
	}
	return nil
}

func registerCommands(config *AppConfig, s *discordgo.Session, logger *zap.SugaredLogger) ([]*discordgo.ApplicationCommand, error) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "minecraft-restart",
			Description: "Minecraftサーバを再起動します",
		},
		{
			Name:        "minecraft-start",
			Description: "Minecraftサーバを起動します（既に起動中の場合は何もしません）",
		},
		{
			Name:        "minecraft-stop",
			Description: "Minecraftサーバを停止します",
		},
		{
			Name:        "minecraft-check",
			Description: "Minecraftサーバの起動状況を確認します",
		},
	}
	commandHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
		"minecraft-restart": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Minecraftサーバを再起動しています...",
				},
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}

			// Stop server if running
			if err := runCommand(config.CheckCmd); err == nil {
				// Server is running
				if err := stopServer(config, logger); err != nil {
					logger.Errorln("Failed to stop server: ", err)
					_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
						Content: "Minecraftサーバの停止に失敗しました",
					})
					if err != nil {
						logger.Errorln("Failed to send message: ", err)
						return
					}
					return
				}
			}

			// Start server
			if err := startServer(config, logger); err != nil {
				logger.Errorln("Failed to start server: ", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Minecraftサーバの起動に失敗しました",
				})
				if err != nil {
					logger.Errorln("Failed to send message: ", err)
					return
				}
				return
			}

			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Minecraftサーバを起動しました",
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}
		},
		"minecraft-start": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Minecraftサーバを起動しています...",
				},
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}

			// Start server if not running
			if err := runCommand(config.CheckCmd); err == nil {
				// Server is running
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Minecraftサーバは既に起動中です",
				})
				if err != nil {
					logger.Errorln("Failed to send message: ", err)
					return
				}
				return
			}

			// Start server
			if err := startServer(config, logger); err != nil {
				logger.Errorln("Failed to start server: ", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Minecraftサーバの起動に失敗しました",
				})
				if err != nil {
					logger.Errorln("Failed to send message: ", err)
					return
				}
				return
			}

			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Minecraftサーバを起動しました",
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}
		},
		"minecraft-stop": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Minecraftサーバを停止しています...",
				},
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}

			// Start server if not running
			if err := runCommand(config.CheckCmd); err != nil {
				// Server is not running
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Minecraftサーバは既に停止しています",
				})
				if err != nil {
					logger.Errorln("Failed to send message: ", err)
					return
				}
				return
			}

			// Stop server
			if err := stopServer(config, logger); err != nil {
				logger.Errorln("Failed to stop server: ", err)
				_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
					Content: "Minecraftサーバの停止に失敗しました",
				})
				if err != nil {
					logger.Errorln("Failed to send message: ", err)
					return
				}
				return
			}

			_, err = s.FollowupMessageCreate(i.Interaction, true, &discordgo.WebhookParams{
				Content: "Minecraftサーバを停止しました",
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}
		},
		"minecraft-check": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
			var content string
			err := runCommand(config.CheckCmd)
			if err := runCommand(config.CheckCmd); err == nil {
				content = "Minecraftサーバは起動しています"
			} else {
				logger.Infoln("Check minecraft: ", err)
				content = "Minecraftサーバは起動していません"
			}
			err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: content,
				},
			})
			if err != nil {
				logger.Errorln("Failed to send message: ", err)
				return
			}
		},
	}

	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := s.ApplicationCommandCreate(s.State.User.ID, config.GuildID, v)
		if err != nil {
			return nil, err
		}
		registeredCommands[i] = cmd
	}
	return registeredCommands, nil
}

func unregisterCommands(config *AppConfig, s *discordgo.Session, commands []*discordgo.ApplicationCommand) error {
	for _, v := range commands {
		err := s.ApplicationCommandDelete(s.State.User.ID, config.GuildID, v.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
