package appconfig

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/xypwn/filediver/config"
)

// This is the central config structure used in the GUI
// and CLI.
// Tags starting with "t:" are custom and used to
// identify which game file types the category
// is targeting.
type Config struct {
	Gamedir string `cfg:"short=g tags=directory default=<auto-detect> help='Helldivers 2 game directory'"`
	Audio   struct {
		Format string `cfg:"options=ogg,wav,aac,mp3,wwise,raw help='common media formats: ogg,wav,aac,mp3; wwise to extract as wem/bnk'"`
	} `cfg:"tags=t:wwise_stream,t:wwise_bank help='audio collections/streams'"`
	Video struct {
		Format string `cfg:"options=mp4,bik,raw"`
	} `cfg:"tags=t:bik help='video streams'"`
	Texture struct {
		Format string `cfg:"options=png,dds,raw"`
	} `cfg:"tags=t:texture"`
	Unit struct {
		SingleFile          bool   `cfg:"help='combine all units into a single blend/glb file'"`
		ImageFormat         string `cfg:"tags=advanced options=png,jpeg"`
		PngCompression      string `cfg:"tags=advanced depends=Unit.ImageFormat=png options=default,none,fast,best"`
		JpegQuality         int    `cfg:"tags=advanced depends=Unit.ImageFormat=jpeg range=1...100 default=90"`
		AllTextures         bool   `cfg:"tags=advanced help='include all referenced textures, including wounds, marks etc. and unknown ones'"`
		AccurateOnly        bool   `cfg:"tags=advanced"`
		SampleAnimations    bool   `cfg:"help='more accurate for now, as spline interpolation conversion isn\\'t implemented yet'"`
		AnimationSampleRate int    `cfg:"depends=Unit.SampleAnimations range=12...144 default=24"`
	} `cfg:"help='general unit settings, affects materials, models and animations'"`
	Material struct {
		Format         string `cfg:"options=blend,glb,textures,raw help='material export format; textures dumps all referenced textures into a folder'"`
		TexturesFormat string `cfg:"depends=Material.Format=textures options=png,dds help='format of individual textures if Format is textures'"`
	} `cfg:"tags=t:material help='see unit options'"`
	Model struct {
		Format                    string `cfg:"options=blend,glb,raw help='model export format'"`
		IncludeLODS               bool   `cfg:"help='include meshes of all levels-of-detail'"`
		EnableAnimations          bool   `cfg:"help='export model animations, can take much longer'"`
		EnableAnimationController bool   `cfg:"tags=advanced depends=Model.EnableAnimations help='export model animation controller, can take even longer and will add many constraints to the output blend file'"`
		JoinComponents            bool   `cfg:"tags=advanced help='join UDIM components'"`
		BoundingBoxes             bool   `cfg:"tags=advanced help='export model bounding boxes'"`
		NoBones                   bool   `cfg:"tags=advanced help='don\\'t include bones'"`
	} `cfg:"tags=t:unit,t:geometry_group,t:prefab help='see unit options'"`
	Animation struct {
		Format string `cfg:"options=json,raw"`
	} `cfg:"tags=t:animation,t:state_machine help='see unit options'"`
	Text struct {
		Format string `cfg:"options=json,raw"`
	} `cfg:"tags=t:strings,t:package,t:bones help='only-text-exportable formats'"`
	Raw struct {
		Format string `cfg:"options=separate,combined,main,stream,gpu help='how to handle the different file sub-types (each file may have a main, stream and GPU file)'"`
	} `cfg:"help='applies to any file without an available extractor or \"raw\" as the selected format'"`
}

// Config must be comparable
var _ = Config{} == Config{}

// Replaces c with preferences in JSON file specified by path.
// Leaves p unchanged if an error occurs. If the file isn't present,
// attempts to write the current state of p to the file.
func (c *Config) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return c.Save(path)
		}
		return err
	}
	newC := *c
	if err := json.Unmarshal(b, &newC); err != nil {
		return err
	}
	*c = newC
	return nil
}

func (c *Config) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path),
		os.ModePerm); err != nil {
		return err
	}
	b, err := json.Marshal(c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0666)
}

var Extractable = map[string]bool{
	"wwise_stream":   true,
	"wwise_bank":     true,
	"bik":            true,
	"material":       true,
	"texture":        true,
	"unit":           true,
	"geometry_group": true,
	"prefab":         true,
	"state_machine":  true,
	"animation":      true,
	"strings":        true,
	"package":        true,
	"bones":          true,
}

var ConfigFields = config.MustFields(Config{})
var _ = GetTypeFormats(Config{}) // config sanity check (see GetTypeFormats)

// GetTypeFormats maps each affected type
// (denoted by t:<type> tag in config struct) to
// the selected format.
// Panics if two different t: tags reference the
// same type.
func GetTypeFormats(extrCfg Config) map[string]string {
	val := reflect.ValueOf(extrCfg)

	formatByType := map[string]string{}
	for _, f := range ConfigFields.Fields {
		if !f.IsCategory {
			continue
		}
		formatFieldName := f.Name + ".Format"
		if _, ok := ConfigFields.ByName[formatFieldName]; !ok {
			continue
		}
		v := val
		for name := range strings.SplitSeq(formatFieldName, ".") {
			v = v.FieldByName(name)
		}
		if v.Kind() != reflect.String {
			continue
		}
		format := v.String()
		for _, tag := range f.Tags {
			if after, ok := strings.CutPrefix(tag, "t:"); ok {
				if _, exists := formatByType[after]; exists {
					panic("GetTypeFormats: game file type " + after + " referenced by multiple categories")
				}
				formatByType[after] = format
			}
		}
	}
	return formatByType
}
