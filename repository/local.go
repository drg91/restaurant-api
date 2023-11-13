package repository

import (
	"encoding/csv"
	"errors"
	"io"
	"net/http"
	"restaurant-api/model"
	"strconv"
	"sync"
	"time"
)

type LocalRepository struct {
	locales    []model.Local
	localesMux sync.Mutex
	lastETag   string
}

func NewLocalRepository(csvURL string) (*LocalRepository, error) {
	repo := &LocalRepository{}
	err := repo.loadCSVData(csvURL)
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *LocalRepository) loadCSVData(csvURL string) error {
	req, err := http.NewRequest(http.MethodGet, csvURL, nil)
	if err != nil {
		return err
	}

	// Agrega el valor de ETag de la solicitud anterior si está disponible
	if r.lastETag != "" {
		req.Header.Set("If-None-Match", r.lastETag)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotModified {
		// No hay cambios en el archivo CSV
		return nil
	} else if resp.StatusCode == http.StatusOK {
		// El archivo ha cambiado, guarda el nuevo valor de ETag
		r.lastETag = resp.Header.Get("ETag")

		reader := csv.NewReader(resp.Body)

		// Omitir la primera línea del CSV si contiene encabezados
		_, err := reader.Read()
		if err != nil {
			return err
		}

		// Lee las líneas del CSV.
		var locales []model.Local
		for {
			line, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				return err
			}

			// Verificar si hay datos de latitud y longitud en la fila
			latitudeStr := line[1]
			longitudeStr := line[2]
			if latitudeStr == "" || longitudeStr == "" {
				continue
			}

			// Parsear los datos del CSV y crear instancias de la estructura Local
			id, err := strconv.Atoi(line[0])
			if err != nil {
				return err
			}

			latitude, err := strconv.ParseFloat(latitudeStr, 64)
			if err != nil {
				return err
			}

			longitude, err := strconv.ParseFloat(longitudeStr, 64)
			if err != nil {
				return err
			}

			availabilityRadius, err := strconv.Atoi(line[3])
			if err != nil {
				return err
			}

			openHour, err := time.Parse("15:04:05", line[4])
			if err != nil {
				return err
			}

			closeHour, err := time.Parse("15:04:05", line[5])
			if err != nil {
				return err
			}

			rating, err := strconv.ParseFloat(line[6], 64)
			if err != nil {
				return err
			}

			local := model.Local{
				ID:                 id,
				Latitude:           latitude,
				Longitude:          longitude,
				AvailabilityRadius: availabilityRadius,
				OpenHour:           openHour,
				CloseHour:          closeHour,
				Rating:             rating,
			}

			locales = append(locales, local)
		}

		// Actualiza la lista de locales con los nuevos datos del CSV
		r.localesMux.Lock()
		r.locales = locales
		r.localesMux.Unlock()

		return nil
	}

	return errors.New("unexpected response from server")
}

func (r *LocalRepository) UpdateCSVData(csvURL string) error {
	r.localesMux.Lock()
	defer r.localesMux.Unlock()

	return r.loadCSVData(csvURL)
}

func (r *LocalRepository) GetLocales() []model.Local {
	r.localesMux.Lock()
	defer r.localesMux.Unlock()

	return r.locales
}
