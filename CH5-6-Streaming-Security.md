# Presigned URLs

Esta implementación fue desarrollada durante la etapa del curso en la que los videos almacenados en el bucket privado de S3 se servían mediante **presigned URLs**.

## `generatePresignedURL`

La función crea un `PresignClient` a partir del cliente S3 y genera una URL temporal para acceder a un objeto privado de S3.

La URL tenía una duración limitada de 15 minutos.

## Cambios realizados en otras partes del proyecto

Para que esta función pudiera funcionar fue necesario modificar otras partes de Tubely:

### `main.go`

Se agregó la configuración del cliente S3 mediante:

* `config.LoadDefaultConfig`
* `config.WithRegion(s3Region)`
* `s3.NewFromConfig(awsCfg)`

El cliente resultante se almacenó en `apiConfig.s3Client`.

También se incorporaron las variables de entorno relacionadas con S3:

* `S3_BUCKET`
* `S3_REGION`

y se agregaron los módulos correspondientes del AWS SDK v2 al proyecto.

### `handler_upload_video.go`

El handler de subida de videos fue modificado para almacenar en `video_url` el bucket y la key del objeto separados por una coma:

```text
bucket,key
```

Por ejemplo:

```text
tubely-private-13467,portrait/abc123.f4v
```

Esto permitía posteriormente recuperar la ubicación del objeto para generar una presigned URL.

### `dbvideo_to_signedvideo.go`

Se agregó `dbVideoToSignedVideo`, que:

1. Comprueba si `VideoURL` es `nil` o está vacío.
2. Separa el bucket y la key almacenados en `video_url`.
3. Llama a `generatePresignedURL`.
4. Reemplaza temporalmente `VideoURL` por la URL firmada.

### `handler_videos_retrieve.go`

La recuperación de videos fue modificada para convertir cada registro de la base de datos mediante `dbVideoToSignedVideo` antes de devolverlo al cliente.

De esta manera, la base de datos conservaba:

```text
bucket,key
```

mientras que la API devolvía una URL temporal utilizable por el frontend.

## Estado actual

Esta implementación fue posteriormente reemplazada por **CloudFront**.

Las funciones se conservan en este archivo `.md` como referencia del proceso de aprendizaje, pero ya no forman parte del código ejecutable de Tubely.


# DB Video → Signed Video

Esta función fue creada para transformar el `video_url` almacenado en la base de datos en una **presigned URL de S3** antes de devolver el video mediante la API.

## Formato almacenado en la base de datos

Durante esta etapa, `handlerUploadVideo` no almacenaba una URL completa. Guardaba:

```text
bucket,key
```

Por ejemplo:

```text
tubely-private-13467,portrait/abc123.f4v
```

## Cambios necesarios para soportarlo

### `database.Video`

`VideoURL` se mantuvo como un `*string`:

```go
VideoURL *string `json:"video_url"`
```

Esto permitía distinguir entre un video que todavía no tenía un archivo asociado (`NULL`) y uno que ya había sido subido.

### `handler_upload_video.go`

Después de subir el archivo a S3, se guardaban el bucket y la key en `VideoURL`:

```text
bucket,key
```

Esto reemplazaba temporalmente el almacenamiento de una URL pública directa.

### `handler_videos_retrieve.go`

Al recuperar los videos del usuario, cada `database.Video` pasaba por:

```go
dbVideoToSignedVideo(video)
```

La función convertía:

```text
bucket,key
```

en:

```text
https://<bucket>.s3.<region>.amazonaws.com/<key>?<presigned parameters>
```

La URL resultante era temporal y permitía acceder al objeto privado sin hacer público el bucket.

## Relación con `generatePresignedURL`

`dbVideoToSignedVideo` utilizaba:

```go
generatePresignedURL(...)
```

para realizar la firma real de la URL.

Por lo tanto, ambas funciones formaban parte de la misma implementación:

```text
database
   │
   │ bucket,key
   ▼
dbVideoToSignedVideo
   │
   ▼
generatePresignedURL
   │
   ▼
presigned S3 URL
```

## Estado actual

Esta implementación fue reemplazada posteriormente por CloudFront.

Se conserva como documentación del ejercicio y como referencia del proceso de aprendizaje.


func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {

    ...
    ...

	videoURL := cfg.s3Bucket + "," + key
	videoMetadata.VideoURL = &videoURL
	if err := cfg.db.UpdateVideo(videoMetadata); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to update video metadata", err)
		return
	}

	videoMetadata, err = cfg.dbVideoToSignedVideo(videoMetadata)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate presigned video URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, videoMetadata)

}


handler video meta funcions:

func (cfg *apiConfig) handlerVideoGet(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid video ID", err)
		return
	}

	video, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusNotFound, "Couldn't get video", err)
		return
	}

	video, err = cfg.dbVideoToSignedVideo(video)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't generate video URL", err)
		return
	}

	respondWithJSON(w, http.StatusOK, video)
}

func (cfg *apiConfig) handlerVideosRetrieve(w http.ResponseWriter, r *http.Request) {
	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}
	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	videos, err := cfg.db.GetVideos(userID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Couldn't retrieve videos", err)
		return
	}

	for i := range videos {
		videos[i], err = cfg.dbVideoToSignedVideo(videos[i])
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't generate video URL", err)
			return
		}
	}

	respondWithJSON(w, http.StatusOK, videos)
}
