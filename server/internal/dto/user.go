package dto

import "mime/multipart"

type UpdateProfileRequest struct {
    Fulname  string  `json:"fullname" binding:"required,min=2,max=100"`
    Gender  *string  `json:"gender"   binding:"required,oneof=MALE FEMALE"`
    Phone   *string  `json:"phone"    binding:"required"`
    Birthday *string `json:"birthday" binding:"required,datetime=2006-01-02"`
}


type UpdateAvatarRequest struct {
	Avatar  *multipart.FileHeader `form:"avatar" validate:"required"`
}

type AddAddressRequest struct {
    Street     string `json:"street" binding:"required"`
    City       string `json:"city" binding:"required"`
    Province   string `json:"province" binding:"required"`
    PostalCode string `json:"postalCode" binding:"required"`
}

type UpdateAddressRequest struct {
    Street     *string `json:"street"`
    City       *string `json:"city"`
    Province   *string `json:"province"`
    PostalCode *string `json:"postalCode"`  
    IsMain     *bool   `json:"isMain"`
}

type DeleteAddressRequest struct {
    AddressID string `json:"addressId" binding:"required"`
}