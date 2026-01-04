package main

import (
	"flag"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfly/flagx"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

var (
	inputFiles = flagx.NewString("input", "", "pattern to match input file(s)")
	verbose    = flagx.NewBool("verbose", false, "verbose logging")
)

func main() {
	generate()
	return
	// protogen.Options{}.Run 会自动处理 stdin/stdout
	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(func(gen *protogen.Plugin) error {
		// 关键点：显式声明支持 proto3 optional
		gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		// 遍历文件进行生成
		/*
			for _, f := range gen.Files {
				if !f.Generate {
					continue
				}
				// generateFile(gen, f) // 你的生成逻辑
			}
		*/
		generate()
		return nil
	})

}

func generate() {
	flagx.Parse()
	var xxxTags string
	var removeTagComment bool

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	if *verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	flagx.Parse()

	var xxxSkipSlice []string
	if len(xxxTags) > 0 {
		log.Info().Msg("warn: deprecated flag '-XXX_skip' used")
		xxxSkipSlice = strings.Split(xxxTags, ",")
	}

	if *inputFiles == "" {
		log.Fatal().Msgf("input file is mandatory, see: -help", os.Args)
	}

	// Note: glob doesn't handle ** (treats as just one *). This will return
	// files and folders, so we'll have to filter them out.
	globResults, err := filepath.Glob(*inputFiles)
	if err != nil {
		log.Fatal().Err(err).Msgf("failed to glob input file %q", inputFiles)
	}

	log.Info().Int("total_files", len(globResults)).Str("input", *inputFiles).Msgf("These files will be processed")

	var matched int
	for _, path := range globResults {
		log.Info().Msgf("info: processing file %s", path)
		finfo, err := os.Stat(path)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to stat file %q", path)
		}

		if finfo.IsDir() {
			continue
		}

		// It should end with ".go" at a minimum.
		if !strings.HasSuffix(strings.ToLower(finfo.Name()), ".go") {
			continue
		}

		matched++

		areas, err := parseFile(path, nil, xxxSkipSlice)
		if err != nil {
			log.Fatal().Err(err).Msgf("failed to parse file %q", path)
		}
		if err = writeFile(path, areas, removeTagComment); err != nil {
			log.Fatal().Err(err).Msgf("failed to write file %q", path)
		}
	}

	if matched == 0 {
		log.Warn().Msgf("input %q matched no files", *inputFiles)
	}
}
