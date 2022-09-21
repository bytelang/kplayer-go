package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/bytelang/kplayer/types"
	"github.com/bytelang/kplayer/types/config"
	"github.com/golang/protobuf/ptypes"
	"github.com/manifoldco/promptui"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"google.golang.org/protobuf/types/known/anypb"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
)

func addInitDefaultCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default",
		Short: "export default config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// init config
			cfg := getDefaultConfig()

			// export file
			return exportConfigFile(cfg, DefaultConfigFileName)
		},
	}

	return cmd
}

func addInitInteractionCommands() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "interaction",
		Short: "interaction init config file",
		RunE: func(cmd *cobra.Command, args []string) error {
			// interaction
			cfg, err := initInteractionConfig()
			if err != nil {
				return err
			}

			// export file
			return exportConfigFile(cfg, DefaultConfigFileName)
		},
	}

	return cmd
}

func exportConfigFile(cfg *config.KPConfig, path string) error {
	d, err := json.Marshal(cfg)
	if err != nil {
		return err
	}

	var indentCfg bytes.Buffer
	if err := json.Indent(&indentCfg, d, "", "  "); err != nil {
		return err
	}

	defer log.WithField("path", path).Info("config file create success")
	return ioutil.WriteFile(path, indentCfg.Bytes(), 0666)
}

func getDefaultConfig() *config.KPConfig {
	res := &config.SingleResource{
		Path: "/video/exmaple.mp4",
	}
	pbs, _ := ptypes.MarshalAny(res)
	return &config.KPConfig{
		Version: ConfigVersion,
		Resource: config.Resource{
			Lists:      []*anypb.Any{pbs},
			Extensions: []string{"mp4", "flv"},
		},
		Play: config.Play{
			StartPoint:          1,
			PlayModel:           strings.ToLower(config.PLAY_MODEL_name[int32(config.PLAY_MODEL_LIST)]),
			EncodeModel:         strings.ToLower(config.ENCODE_MODEL_name[int32(config.ENCODE_MODEL_RTMP)]),
			CacheOn:             false,
			CacheUncheck:        false,
			SkipInvalidResource: false,
			Rpc: &config.Server{
				On:       true,
				Address:  types.DefaultRPCAddress,
				GrpcPort: types.DefaultRPCPort,
				HttpPort: types.DefaultHttpPort,
			},
			Encode: &config.Encode{
				VideoWidth:         780,
				VideoHeight:        480,
				VideoFps:           30,
				AudioSampleRate:    48000,
				AudioChannelLayout: 3,
				AudioChannels:      2,
				BitRate:            0,
				AvgQuality:         0,
			},
		},
		Output: config.Output{
			ReconnectInternal: -1,
			Lists: func() (list []*config.OutputInstance) {
				list = append(list, &config.OutputInstance{
					Path:   "rtmp://127.0.0.1:1935/live",
					Unique: "test_output",
				})
				return
			}(),
		},
		Plugin: config.Plugin{
			Lists: []*config.PluginInstance{},
		},
	}
}

type interActionContent struct {
	Validator func(string) error
	Label     string
	Func      func(line string) error
}

func initInteractionConfig() (*config.KPConfig, error) {
	cfg := getDefaultConfig()

	interActions := []interActionContent{}
	interActions = append(interActions, interActionContent{
		Label: "Which directory of resource files do you want to read?",
		Validator: func(line string) error {
			s, err := os.Stat(line)
			if err != nil {
				return err
			}
			if !s.IsDir() {
				return fmt.Errorf("Please input directory path")
			}
			return nil
		},
		Func: func(line string) error {
			fileInfo, err := ioutil.ReadDir(line)
			if err != nil {
				return err
			}
			allowExtension := map[string]bool{".mp4": true, ".flv": true, ".mkv": true, ".rmvb": true, ".avi": true, ".3gp": true, ".hevc": true}
			for _, item := range fileInfo {
				if !item.IsDir() {
					filePath := item.Name()
					if _, ok := allowExtension[path.Ext(filePath)]; ok {
						path.Join()
						res := &config.SingleResource{
							Path: path.Join(line, filePath),
						}
						pbs, _ := ptypes.MarshalAny(res)
						cfg.Resource.Lists = append(cfg.Resource.Lists, pbs)
					}
				}
			}
			return nil
		},
	})
	interActions = append(interActions, interActionContent{
		Label: "Which number do you want to start with? [default: 0]",
		Func: func(line string) error {
			if line == "" {
				line = "0"
			}
			n, err := strconv.ParseUint(line, 10, 32)
			if err != nil {
				return fmt.Errorf("input number must be integer")
			}
			cfg.Play.StartPoint = uint32(n)
			return nil
		},
	})
	interActions = append(interActions, interActionContent{
		Label: "Whether to open jsonrpc yes/no? [default: yes]",
		Func: func(line string) error {
			if line == "no" {
				cfg.Play.Rpc.On = false
			}
			return nil
		},
	})
	interActions = append(interActions, interActionContent{
		Label: "Please enter the output file path or rtmp server address",
		Func: func(line string) error {
			if line == "" {
				return fmt.Errorf("output path cannot be empty")
			}
			outputInstances := &config.OutputInstance{
				Path: line,
			}
			cfg.Output.Lists = append(cfg.Output.Lists, outputInstances)
			return nil
		},
	})

	for _, item := range interActions {
		prompt := promptui.Prompt{
			Label:  item.Label,
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
		}
		result, err := prompt.Run()
		if err != nil {
			return nil, err
		}

		if err := item.Func(result); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}
