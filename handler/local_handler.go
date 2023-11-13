package handler

import (
	"encoding/json"
	"log"
	"math"
	"net/http"
	"restaurant-api/model"
	"restaurant-api/repository"
	"strconv"
	"sync"
	"time"
)

const R = 6371.01

type LocalHandler struct {
	localRepo *repository.LocalRepository
}

func NewLocalHandler(localRepo *repository.LocalRepository) *LocalHandler {
	return &LocalHandler{localRepo: localRepo}
}

func (h *LocalHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
	// Obtiene los parámetros de consulta.
	latStr := r.URL.Query().Get("latitude")
	lonStr := r.URL.Query().Get("longitude")

	// Verifica si los parámetros están presentes.
	if latStr == "" || lonStr == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Convierte las cadenas de latitud y longitud a float64.
	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Obtiene la lista de locales del repositorio.
	locales := h.localRepo.GetLocales()

	// Utiliza un canal para recibir los resultados de las goroutines.
	resultChannel := make(chan model.Local)
	var wg sync.WaitGroup

	// Divide la búsqueda en partes y ejecuta goroutines concurrentemente.
	numParts := 10 // Número de partes para dividir la búsqueda
	partSize := len(locales) / numParts
	timeNow := time.Now()

	for i := 0; i < numParts; i++ {
		wg.Add(1)
		go func(start, end int) {
			defer wg.Done()
			for _, local := range locales[start:end] {
				if h.isInRange(lat, lon, local) && h.isOpen(timeNow, local) {
					resultChannel <- local
				}
			}
		}(i*partSize, (i+1)*partSize)
	}

	// Cierra el canal cuando todas las goroutines hayan terminado.
	go func() {
		wg.Wait()
		close(resultChannel)
	}()

	// Recolecta los resultados de las goroutines.
	var results []model.Local
	for local := range resultChannel {
		results = append(results, local)
	}
	log.Println("Cantidad de locales encontrados: ", len(results))
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func (h *LocalHandler) isInRange(lat, lon float64, local model.Local) bool {
	// Calcula la distancia entre el punto y el local.
	distance := distanceHaversine(lat, lon, local.Latitude, local.Longitude)

	// Verifica si la distancia es menor o igual al radio de alcance del local.
	return distance <= float64(local.AvailabilityRadius)
}

func (h *LocalHandler) isOpen(now time.Time, local model.Local) bool {
	nowHour, nowMin, nowSec := now.Clock()
	openHour, openMin, openSec := local.OpenHour.Clock()
	closeHour, closeMin, closeSec := local.CloseHour.Clock()

	nowTimeInSeconds := nowHour*3600 + nowMin*60 + nowSec
	openTimeInSeconds := openHour*3600 + openMin*60 + openSec
	closeTimeInSeconds := closeHour*3600 + closeMin*60 + closeSec

	return nowTimeInSeconds > openTimeInSeconds && nowTimeInSeconds < closeTimeInSeconds
}

func distanceEquirectangular(lat1, lon1, lat2, lon2 float64) float64 {
	x := (lon2 - lon1) * math.Cos((lat1+lat2)/2)
	y := lat2 - lat1
	return R * math.Sqrt(x*x+y*y)
}

func distanceLeyCosenos(lat1, lon1, lat2, lon2 float64) float64 {
	// Convertir grados a radianes
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Aplicar la fórmula de Ley de Cosenos Esférica
	return math.Acos(math.Sin(lat1Rad)*math.Sin(lat2Rad)+math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Cos(lon2Rad-lon1Rad)) * R
}

func distanceHaversine(lat1, lon1, lat2, lon2 float64) float64 {
	// Convertir grados a radianes
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Diferencia de las coordenadas
	dLat := lat2Rad - lat1Rad
	dLon := lon2Rad - lon1Rad

	// Aplicar la fórmula de Haversine
	a := math.Pow(math.Sin(dLat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dLon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Distancia en kilómetros
	return R * c
}
