package transport

import (
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

const (
	attrImg    = "image"
	attrIsMain = "is_main_image"
)

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
	return r
}

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

	artIdVal := chi.URLParam(r, ArtworkIDParam)
	artworkID, err := uuid.Parse(artIdVal)
	if err != nil {
		return nil, ErrInvalidUUID
	}

	isMainVal := r.FormValue(attrIsMain)
	isMainImage, err := strconv.ParseBool((isMainVal))
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

func handleImgServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrUnsupportedFormat):
		utils.RespondError(w, http.StatusBadRequest, "unsupported image format")
	default:
		log.Printf("image service error: %v", err)
		utils.RespondServerError(w)
	}
}
