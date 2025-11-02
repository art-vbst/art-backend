package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/art-vbst/art-backend/internal/artwork/domain"
	"github.com/art-vbst/art-backend/internal/artwork/repo"
	"github.com/art-vbst/art-backend/internal/artwork/service"
	"github.com/art-vbst/art-backend/internal/platform/db/store"
	"github.com/art-vbst/art-backend/internal/platform/storage"
	"github.com/art-vbst/art-backend/internal/platform/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

const ArtworkIDParam = "artworkID"

var (
	ErrInvalidUUID     = errors.New("invalid UUID")
	ErrInvalidFormData = errors.New("invalid value")
)

type ImageHandler struct {
	service *service.ImageService
}

func NewImageHandler(db *store.Store, provider storage.Provider) *ImageHandler {
	service := service.NewImageService(repo.New(db), provider)
	return &ImageHandler{service: service}
}

func (h *ImageHandler) Routes() chi.Router {
	r := chi.NewRouter()
	r.Post("/", h.create)
	r.Put("/{id}", h.update)
	r.Delete("/{id}", h.delete)
	return r
}

const (
	attrImg    = "image"
	attrIsMain = "is_main_image"
)

func (h *ImageHandler) create(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r); err != nil {
		return
	}

	data, err := h.parseCreateRequest(r)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidUUID):
			utils.RespondError(w, http.StatusBadRequest, "invalid artwork id")
		case errors.Is(err, ErrInvalidFormData):
			utils.RespondError(w, http.StatusBadRequest, "invalid form data")
		default:
			handleImgServiceError(w, err)
		}
		return
	}
	defer data.File.Close()

	image, err := h.service.Create(r.Context(), data)
	if err != nil {
		handleImgServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, image)
}

func (h *ImageHandler) parseCreateRequest(r *http.Request) (*service.CreateImageData, error) {
	r.ParseMultipartForm(10 * utils.MB)

	artworkID, err := uuid.Parse(chi.URLParam(r, ArtworkIDParam))
	if err != nil {
		return nil, ErrInvalidUUID
	}

	isMainVal := r.FormValue(attrIsMain)
	isMainImage, err := strconv.ParseBool(isMainVal)
	if err != nil {
		return nil, ErrInvalidFormData
	}

	file, fileHeader, err := r.FormFile(attrImg)
	if err != nil {
		return nil, ErrInvalidFormData
	}

	width, height, err := h.service.GetImageDimensions(file)
	if err != nil {
		return nil, err
	}

	return &service.CreateImageData{
		UploadFileData: storage.UploadFileData{
			File:        file,
			FileName:    fileHeader.Filename,
			ContentType: fileHeader.Header.Get("Content-Type"),
		},
		CreateImagePayload: domain.CreateImagePayload{
			ArtworkID:   artworkID,
			IsMainImage: isMainImage,
			ImageWidth:  width,
			ImageHeight: height,
		},
	}, nil
}

type updatePayload struct {
	IsMainImage string `json:"is_main_image"`
}

func (h *ImageHandler) update(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r); err != nil {
		return
	}

	id, idErr := uuid.Parse(chi.URLParam(r, "id"))
	artID, artIDErr := uuid.Parse(chi.URLParam(r, ArtworkIDParam))
	if idErr != nil || artIDErr != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad uuid")
	}

	var body updatePayload
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	isMainImage, err := strconv.ParseBool(body.IsMainImage)
	if err != nil {
		utils.RespondError(w, http.StatusBadRequest, "Invalid request body")
	}

	img, err := h.service.Update(r.Context(), artID, id, isMainImage)
	if err != nil {
		handleImgServiceError(w, err)
		return
	}

	utils.RespondJSON(w, http.StatusOK, img)
}

func (h *ImageHandler) delete(w http.ResponseWriter, r *http.Request) {
	if _, err := utils.Authenticate(w, r); err != nil {
		return
	}

	id, idErr := uuid.Parse(chi.URLParam(r, "id"))
	artID, artIDErr := uuid.Parse(chi.URLParam(r, ArtworkIDParam))
	if idErr != nil || artIDErr != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad uuid")
	}

	err := h.service.Delete(r.Context(), artID, id)
	if err != nil {
		handleImgServiceError(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleImgServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUnsupportedFormat):
		utils.RespondError(w, http.StatusBadRequest, "unsupported image format")
	case errors.Is(err, sql.ErrNoRows):
	case errors.Is(err, service.ErrInvalidArtID):
		utils.RespondError(w, http.StatusNotFound, "artwork or image not found")
	default:
		log.Printf("image service error: %v", err)
		utils.RespondServerError(w)
	}
}
