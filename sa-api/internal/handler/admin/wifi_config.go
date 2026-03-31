package admin

import (
	"strconv"

	"github.com/hdbank/smart-attendance/internal/domain/entity"
	"github.com/hdbank/smart-attendance/internal/domain/repository"
	"github.com/hdbank/smart-attendance/pkg/apperrors"
	"github.com/hdbank/smart-attendance/pkg/response"
	"github.com/labstack/echo/v4"
)

// WiFiConfigHandler xử lý các API quản lý cấu hình WiFi của chi nhánh
type WiFiConfigHandler struct {
	wifiConfigRepo repository.WiFiConfigRepository
}

// NewWiFiConfigHandler tạo instance WiFiConfigHandler
func NewWiFiConfigHandler(wifiConfigRepo repository.WiFiConfigRepository) *WiFiConfigHandler {
	return &WiFiConfigHandler{wifiConfigRepo: wifiConfigRepo}
}

// CreateRequest yêu cầu tạo WiFi config
type CreateWiFiConfigRequest struct {
	SSID        string `json:"ssid"`
	BSSID       string `json:"bssid"`
	Description string `json:"description"`
}

// UpdateRequest yêu cầu cập nhật WiFi config
type UpdateWiFiConfigRequest struct {
	SSID        *string `json:"ssid"`
	BSSID       *string `json:"bssid"`
	Description *string `json:"description"`
	IsActive    *bool   `json:"is_active"`
}

// GetByBranch godoc
// @Summary Lấy danh sách WiFi config của chi nhánh
// @Tags Admin - WiFi Config
// @Security BearerAuth
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Success 200 {object} response.Response{data=[]entity.WiFiConfig}
// @Router /admin/branches/{branch_id}/wifi-configs [get]
func (h *WiFiConfigHandler) GetByBranch(c echo.Context) error {
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	configs, err := h.wifiConfigRepo.FindByBranch(c.Request().Context(), uint(branchID))
	if err != nil {
		return response.Error(c, err)
	}

	return response.OK(c, configs)
}

// Create godoc
// @Summary Thêm WiFi config cho chi nhánh (Admin)
// @Tags Admin - WiFi Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param body body CreateWiFiConfigRequest true "Thông tin WiFi"
// @Success 201 {object} response.Response{data=entity.WiFiConfig}
// @Router /admin/branches/{branch_id}/wifi-configs [post]
func (h *WiFiConfigHandler) Create(c echo.Context) error {
	branchID, err := strconv.ParseUint(c.Param("branch_id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	var req CreateWiFiConfigRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if req.SSID == "" {
		return response.Error(c, apperrors.NewValidationError(map[string]string{
			"ssid": "SSID không được để trống",
		}))
	}

	config := &entity.WiFiConfig{
		BranchID:    uint(branchID),
		SSID:        req.SSID,
		BSSID:       req.BSSID,
		Description: req.Description,
		IsActive:    true,
	}

	if err := h.wifiConfigRepo.Create(c.Request().Context(), config); err != nil {
		return response.Error(c, err)
	}

	return response.Created(c, config)
}

// Update godoc
// @Summary Cập nhật WiFi config (Admin)
// @Tags Admin - WiFi Config
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param id path int true "WiFi Config ID"
// @Param body body UpdateWiFiConfigRequest true "Thông tin cập nhật"
// @Success 200 {object} response.Response{data=entity.WiFiConfig}
// @Router /admin/branches/{branch_id}/wifi-configs/{id} [put]
func (h *WiFiConfigHandler) Update(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	config, err := h.wifiConfigRepo.FindByID(c.Request().Context(), uint(id))
	if err != nil {
		return response.Error(c, err)
	}

	var req UpdateWiFiConfigRequest
	if err := c.Bind(&req); err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	// Chỉ cập nhật field có data (partial update)
	if req.SSID != nil {
		config.SSID = *req.SSID
	}
	if req.BSSID != nil {
		config.BSSID = *req.BSSID
	}
	if req.Description != nil {
		config.Description = *req.Description
	}
	if req.IsActive != nil {
		config.IsActive = *req.IsActive
	}

	if err := h.wifiConfigRepo.Update(c.Request().Context(), config); err != nil {
		return response.Error(c, err)
	}

	return response.OKWithMessage(c, "Cập nhật WiFi config thành công", config)
}

// Delete godoc
// @Summary Xóa WiFi config (Admin)
// @Tags Admin - WiFi Config
// @Security BearerAuth
// @Produce json
// @Param branch_id path int true "Branch ID"
// @Param id path int true "WiFi Config ID"
// @Success 204
// @Router /admin/branches/{branch_id}/wifi-configs/{id} [delete]
func (h *WiFiConfigHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return response.Error(c, apperrors.ErrValidation)
	}

	if err := h.wifiConfigRepo.Delete(c.Request().Context(), uint(id)); err != nil {
		return response.Error(c, err)
	}

	return response.NoContent(c)
}
