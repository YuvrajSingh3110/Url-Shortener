package routes

import (
	"os"
	"strconv"
	"time"

	"github.com/YuvrajSingh3110/Url_Shortener/database"
	"github.com/YuvrajSingh3110/Url_Shortener/helpers"
	"github.com/asaskevich/govalidator"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/go-redis/redis/v8"
)

type Request struct {
	Url         string        `json:"url"`
	CustomShort string        `json:"short"`
	Expiry      time.Duration `json:"expiry"`
}

type Response struct {
	Url            string        `json:"url"`
	CustomShort    string        `json:"short"`
	Expiry         time.Duration `json:"expiry"`
	XRateRemaining int           `json:"rate_limit"`
	XRateReset     time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {
	body := new(Request)
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse JSON"})
	}

	//rate limiting
	r2 := database.CreateClient(1)
	defer r2.Close()

	//we check for ip address
	val, err := r2.Get(database.Ctx, c.IP()).Result()

	//if u didn't find any value in the db which means user is using the api for the 1st time
	if err == redis.Nil {
		_ = r2.Set(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), 30*60*time.Second).Err()
	} else { //user found
		val, _ = r2.Get(database.Ctx, c.IP()).Result()
		valInt, _ := strconv.Atoi(val)
		if valInt <= 0 {
			limit, _ := r2.TTL(database.Ctx, c.IP()).Result()
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error":            "rate limit exceeded",
				"rate_limit_reset": limit / time.Nanosecond / time.Minute,
			})
		}
	}

	//check if input is an url
	if !govalidator.IsURL(body.Url) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid URL"})
	}

	//checking for domain error to prevent infinite loop
	if !helpers.RemoveDomainError(body.Url) {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "you cannot hack this system :)"})
	}

	//enforce https, SSL
	body.Url = helpers.EnforceHTTP(body.Url)

	var id string
	if body.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = body.CustomShort
	}

	r := database.CreateClient(0)
	defer r.Close()
	val, _ = r.Get(database.Ctx, id).Result()
	if val != "" {
		//something is found with that uid
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "url custom is already in use",
		})
	}

	if body.Expiry == 0 {
		body.Expiry = 24 //deafult expiry of 24 hours
	}

	err = r.Set(database.Ctx, id, body.Url, body.Expiry*3600*time.Second).Err()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Unable to connec Server",
		})
	}

	//sending the reponse
	resp := Response{
		Url:            body.Url,
		CustomShort:    "",
		Expiry:         body.Expiry,
		XRateRemaining: 10,
		XRateReset:     30,
	}

	r2.Decr(database.Ctx, c.IP())

	val, _ = r2.Get(database.Ctx, c.IP()).Result()
	resp.XRateRemaining, _ = strconv.Atoi(val)

	ttl, _ := r2.TTL(database.Ctx, c.IP()).Result()
	resp.XRateReset = ttl / time.Nanosecond / time.Minute

	resp.CustomShort = os.Getenv("DOMAIN") + "/" + id

	return c.Status(fiber.StatusOK).JSON(resp)
}
