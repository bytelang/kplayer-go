package app

import (
    "fmt"
    "github.com/bytelang/kplayer/types"
    "github.com/bytelang/kplayer/types/config"
    "github.com/manifoldco/promptui"
    "io/ioutil"
    "os"
    "path"
    "strconv"
)

func getDefaultConfig() *config.KPConfig {
    return &config.KPConfig{
        Version: ConfigVersion,
        Resource: config.Resource{
            Lists: []string{},
        },
        Play: config.Play{
            PlayModel:   config.PlayModel_list,
            EncodeModel: config.EncodeModel_file,
            Encode: config.Encode{
                VideoWidth:      780,
                VideoHeight:     480,
                VideoFps:        30,
                AudioSampleRate: 48000,
            },
            Jsonrpc: false,
        },
        Output: config.Output{
            Lists: []config.OutputInstance{},
        },
        Plugin: config.Plugin{
            Lists: []config.PluginInstance{},
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
                        cfg.Resource.Lists = append(cfg.Resource.Lists, path.Join(line, filePath))
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
        Label: "Whether to open jsonrpc yes/no? [default: no]",
        Func: func(line string) error {
            if line == "yes" {
                cfg.Play.Jsonrpc = true
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
            outputInstances := config.OutputInstance{
                Path:   line,
                Unique: types.GetRandString(6),
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
