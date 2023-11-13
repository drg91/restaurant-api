package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"restaurant-api/handler"
	"restaurant-api/repository"
	"time"
)

func main() {
	// Obtiene la URL del CSV de una variable de entorno
	csvURL := os.Getenv("CSV_URL")
	if csvURL == "" {
		log.Fatalf("Error: La variable de entorno CSV_URL no está definida")
	}

	localRepo, err := repository.NewLocalRepository(csvURL)
	if err != nil {
		// Maneja el error si la inicialización del repositorio falla
		panic(err)
	}

	localHandler := handler.NewLocalHandler(localRepo)
	http.HandleFunc("/", localHandler.HandleRequest)

	// Inicia la goroutine para actualizar los datos del CSV cada 5 minutos
	go func() {
		for {
			inicio := time.Now()
			// Actualiza los datos del CSV
			err := localRepo.UpdateCSVData(csvURL)
			if err != nil {
				log.Println("Error al cargar datos del CSV:", err)
			} else {
				log.Println("Datos del CSV actualizados correctamente.")
			}
			duracion := time.Since(inicio) // Calcula la duración desde el inicio

			log.Printf("La función tardó %v en ejecutarse\n", duracion)
			// Espera 10 minutos antes de la próxima actualización
			time.Sleep(10 * time.Minute)
		}
	}()
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}

}
