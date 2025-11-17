package main

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/net/html/charset"
	"gopkg.in/yaml.v3"
)

type Config struct {
	InputFile  string `yaml:"input-file"`
	OutputFile string `yaml:"output-file"`
}

type FloatValue float64

func (f *FloatValue) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var stringValue string

	err := d.DecodeElement(&stringValue, &start)
	if err != nil {
		return fmt.Errorf("decode value: %w", err)
	}

	stringValue = strings.ReplaceAll(stringValue, ",", ".")

	value, err := strconv.ParseFloat(stringValue, 64)
	if err != nil {
		return fmt.Errorf("parse float in value: %w", err)
	}

	*f = FloatValue(value)

	return nil
}

type Record struct {
	ID    int       `json:"num_code"  xml:"NumCode"`
	Name  string    `json:"char_code" xml:"CharCode"`
	Value FloatValue `json:"value"     xml:"Value"`
}

type RawRecords struct {
	Items []Record `xml:"Valute"`
}

func parseConfig() (Config, error) {
	configPath := flag.String("config", "config.yaml", "Path to the configuration file")
	flag.Parse()

	data, err := os.ReadFile(*configPath)
	if err != nil {
		return Config{}, fmt.Errorf("read yaml: %w", err)
	}

	var cfg Config

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return Config{}, fmt.Errorf("decode yaml: %w", err)
	}

	return cfg, nil
}

func parseXML(path string) ([]Record, error) {
	xmlData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read xml file %s: %w", path, err)
	}

	var raw RawRecords

	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	decoder.CharsetReader = charset.NewReaderLabel

	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode xml: %w", err)
	}

	return raw.Items, nil
}

func saveAsJSON(items []Record, path string) error {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Value > items[j].Value
	})

	dir := filepath.Dir(path)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("create dir %s: %w", dir, err)
	}

	jsonData, err := json.MarshalIndent(items, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json %w", err)
	}

	err = os.WriteFile(path, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("write file %s: %w", path, err)
	}

	return nil
}

func main() {
	cfg, err := parseConfig()
	if err != nil {
		panic(err)
	}

	records, err := parseXML(cfg.InputFile)
	if err != nil {
		panic(err)
	}

	err = saveAsJSON(records, cfg.OutputFile)
	if err != nil {
		panic(err)
	}
}