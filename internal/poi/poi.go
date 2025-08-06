package poi

import (
	"encoding/csv"
	"strconv"
)

type POI struct {
	Lon         float64
	Lat         float64
	Color       string
	Text        string
	FontSize    int32
	MaxLod      int32
	Transparent bool
	Demand      string
	Population  int64
}

type List []POI

func (p *List) Add(poi POI) {
	*p = append(*p, poi)
}

func (p *List) ToTSV(w *csv.Writer) error {
	w.Comma = '\t'
	header := []string{"lon", "lat", "color", "text", "font_size", "max_lod", "transparent", "demand", "population"}
	if err := w.Write(header); err != nil {
		return err
	}
	for _, poi := range *p {
		record := []string{
			strconv.FormatFloat(poi.Lon, 'f', -1, 64),
			strconv.FormatFloat(poi.Lat, 'f', -1, 64),
			poi.Color,
			poi.Text,
			strconv.Itoa(int(poi.FontSize)),
			strconv.Itoa(int(poi.MaxLod)),
			strconv.FormatBool(poi.Transparent),
			poi.Demand,
			strconv.FormatInt(poi.Population, 10),
		}
		if err := w.Write(record); err != nil {
			return err
		}
	}
	return nil
}
