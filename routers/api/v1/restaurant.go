package v1

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/punchanabu/redrice-backend-go/models"
	"github.com/punchanabu/redrice-backend-go/utils"
	"gorm.io/gorm"
)

var RestaurantHandler *models.RestaurantHandler

func InitializedRestaurantHandler(db *gorm.DB) {
	RestaurantHandler = models.NewRestaurantHandler(db)
}

// @Summary Get a Single Restaurant
// @Description Retrieves details of a single restaurant by its unique identifier.
// @Tags restaurants
// @Produce json
// @Param id path int true "Restaurant ID" Format(int64)
// @security BearerAuth
// @Success 200 {object} models.Restaurant "The details of the restaurant including ID, name, location, and other relevant information."
// @Failure 400 {object} ErrorResponse "Invalid restaurant ID format."
// @Failure 404 {object} ErrorResponse "Restaurant not found with the specified ID."
// @Router /restaurants/{id} [get]
func GetRestaurant(c *gin.Context) {
	idString := c.Param("id")
	idInt, err := strconv.Atoi(idString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant id"})
		return
	}

	idUint := uint(idInt)
	restaurant, err := RestaurantHandler.GetRestaurant(idUint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, restaurant)
}

// @Summary Get All Restaurants
// @Description Retrieves a list of all restaurants in the system.
// @Tags restaurants
// @Produce json
// @security BearerAuth
// @Success 200 {array} models.Restaurant "An array of restaurant objects."
// @Failure 500 {object} ErrorResponse "Internal server error while fetching restaurants."
// @Router /restaurants [get]
func GetRestaurants(c *gin.Context) {
	users, err := RestaurantHandler.GetRestaurants()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error fetching restaurants!"})
		return
	}
	c.JSON(http.StatusOK, users)
}

// @Summary Create a New Restaurant
// @Description Adds a new restaurant to the system with the provided details.
// @Tags restaurants
// @Accept json
// @Produce json
// @Param restaurant body models.Restaurant true "Restaurant Registration Details"
// @security BearerAuth
// @Success 201 {object} models.Restaurant "The created restaurant's details, including its unique identifier."
// @Failure 400 {object} ErrorResponse "Invalid input format for restaurant details."
// @Failure 500 {object} ErrorResponse "Internal server error while creating the restaurant."
// @Router /restaurants [post]
func CreateRestaurant(c *gin.Context) {

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form!"})
		return
	}

	name := c.Request.FormValue("name")
	address := c.Request.FormValue("address")
	telephone := c.Request.FormValue("telephone")
	description := c.Request.FormValue("description")
	facebook := c.Request.FormValue("facebook")
	instagram := c.Request.FormValue("instagram")
	openTime := c.Request.FormValue("openTime")
	closeTime := c.Request.FormValue("closeTime")
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing image!"})
		return
	}

	defer file.Close()

	imageUrl, err := utils.UploadImageToS3("redrice", file, header.Filename)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading image!"})
		return
	}

	restaurant := models.Restaurant{
		Name:        name,
		Address:     address,
		Telephone:   telephone,
		Description: description,
		ImageURL:    imageUrl,
		Facebook:    facebook,
		Instagram:   instagram,
		OpenTime:    openTime,
		CloseTime:   closeTime,
	}

	if err := RestaurantHandler.CreateRestaurant(&restaurant); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating restaurant!"})
		return
	}

	c.JSON(http.StatusCreated, restaurant)
}

// @Summary Update a Restaurant
// @Description Updates the details of an existing restaurant identified by its ID.
// @Tags restaurants
// @Accept json
// @Produce json
// @Param id path int true "Restaurant ID" Format(int64)
// @Param restaurant body models.Restaurant true "Updated Restaurant Details"
// @security BearerAuth
// @Success 200 {object} models.Restaurant "The updated restaurant's details."
// @Failure 400 {object} ErrorResponse "Invalid input format for restaurant details or invalid restaurant ID."
// @Failure 404 {object} ErrorResponse "Restaurant not found with the specified ID."
// @Router /restaurants/{id} [put]
func UpdateRestaurant(c *gin.Context) {
	idString := c.Param("id")
	idInt, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant ID"})
		return
	}
	idUint := uint(idInt)

	// Parse multipart form
	if err := c.Request.ParseMultipartForm(10 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error parsing form"})
		return
	}

	// Extract data from the form
	name := c.Request.FormValue("name")
	address := c.Request.FormValue("address")
	telephone := c.Request.FormValue("telephone")
	description := c.Request.FormValue("description")
	facebook := c.Request.FormValue("facebook")
	instagram := c.Request.FormValue("instagram")
	openTime := c.Request.FormValue("openTime")
	closeTime := c.Request.FormValue("closeTime")
	ratingStr := c.Request.FormValue("rating")
	commentCountStr := c.Request.FormValue("commentCount")

	file, header, err := c.Request.FormFile("image")
	var imageUrl string
	if err == nil {
		defer file.Close()
		imageUrl, err = utils.UploadImageToS3("redrice", file, header.Filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error uploading image"})
			return
		}
	} else {
		imageUrl = ""
	}

	// Create an updated restaurant model
	updatedRestaurant := models.Restaurant{
		Name:        name,
		Address:     address,
		Telephone:   telephone,
		Description: description,
		Facebook:    facebook,
		Instagram:   instagram,
		OpenTime:    openTime,
		CloseTime:   closeTime,
	}
	if imageUrl != "" {
		updatedRestaurant.ImageURL = imageUrl
	}

	if ratingStr != "" {
		rating, err := strconv.ParseFloat(ratingStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rating"})
			return
		}

		updatedRestaurant.Rating = &rating
	}

	if commentCountStr != "" {
		commentCount, err := strconv.ParseFloat(commentCountStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid comment count"})
			return
		}

		updatedRestaurant.CommentCount = &commentCount
	}

	// Update the restaurant in the database
	err = RestaurantHandler.UpdateRestaurant(idUint, &updatedRestaurant)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, updatedRestaurant)
}

// @Summary Delete a Restaurant
// @Description Removes a restaurant from the system by its unique identifier.
// @Tags restaurants
// @Produce json
// @Param id path int true "Restaurant ID" Format(int64)
// @security BearerAuth
// @Success 204 "Restaurant successfully deleted, no content to return."
// @Failure 400 {object} ErrorResponse "Invalid restaurant ID format."
// @Failure 404 {object} ErrorResponse "Restaurant not found with the specified ID."
// @Router /restaurants/{id} [delete]
func DeleteRestaurant(c *gin.Context) {
	idString := c.Param("id")
	idInt, err := strconv.Atoi(idString)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid restaurant id"})
		return
	}

	idUint := uint(idInt)

	err = RestaurantHandler.DeleteRestaurant(idUint)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Restaurant not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "message": "Restaurant deleted successfully!"})
}
