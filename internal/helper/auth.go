package helper

import (
	"errors"
	"fmt"
	"go-ecommerce-app/internal/domain"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Auth struct {
	Secret string
}

func SetupAuth(secret string) Auth {
	if secret == "" {
		panic("JWT_SECRET cannot be empty")
	}
	return Auth{Secret: secret}
}

func (a Auth) CreateHashedPassword(password string) (string, error) {

	if len(password) < 8 {
		return "", errors.New("password must be at least 8 characters long")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		//log actual error and report to logging tool
		return "", errors.New("failed to hash password")
	}
	return string(hashedPassword), nil
}

func (a Auth) GenerateToken(userId uint, email string, role string) (string, error) {

	if userId == 0 || email == "" || role == "" {
		return "", errors.New("invalid user id, email or role")
	}

	if a.Secret == "" {
		return "", errors.New("JWT secret is not configured")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   userId,
		"email": email,
		"role":  role,
		"exp":   time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString([]byte(a.Secret))
	if err != nil {
		return "", errors.New("failed to generate token")
	}

	return tokenString, nil
}

func (a Auth) VerifyPassword(password string, hashedPassword string) (bool, error) {

	if len(password) < 8 {
		return false, errors.New("password must be at least 8 characters long")
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return false, errors.New("invalid password")
	}

	return true, nil
}

func (a Auth) VerifyToken(token string) (domain.User, error) {

	tokenArray := strings.Split(token, " ")

	if len(tokenArray) != 2 {
		return domain.User{}, errors.New("invalid token")
	}

	tokenString := tokenArray[1]

	if tokenArray[0] != "Bearer" {
		return domain.User{}, errors.New("invalid token")
	}
	if a.Secret == "" {
		return domain.User{}, errors.New("JWT secret is not configured")
	}

	parsedToken, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.Secret), nil
	})

	if err != nil {
		return domain.User{}, fmt.Errorf("failed to parse token: %v", err)
	}

	if claims, ok := parsedToken.Claims.(jwt.MapClaims); ok && parsedToken.Valid {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			return domain.User{}, errors.New("token is expired")
		}

		user := domain.User{}
		user.ID = uint(claims["sub"].(float64))
		user.Email = claims["email"].(string)
		user.UserType = claims["role"].(string)

		return user, nil
	}

	return domain.User{}, errors.New("invalid token claims")
}

func (a Auth) Authorize(ctx *fiber.Ctx) error {
	authHeader := ctx.Get("Authorization")
	if authHeader == "" {
		return ctx.Status(401).JSON(fiber.Map{
			"message": "authorization failed",
			"reason":  "Authorization header is missing",
		})
	}

	user, err := a.VerifyToken(authHeader)
	if err == nil && user.ID > 0 {
		ctx.Locals("user", user)
		return ctx.Next()
	} else {
		errMsg := "unauthorized"
		if err != nil {
			errMsg = err.Error()
		}
		return ctx.Status(401).JSON(fiber.Map{
			"message": "authorization failed",
			"reason":  errMsg,
		})
	}
}

func (a Auth) GetCurrentUser(ctx *fiber.Ctx) domain.User {
	user, ok := ctx.Locals("user").(domain.User)
	if !ok {
		return domain.User{}
	}
	return user
}

func (a Auth) GenerateVerificationCode() (int, error) {
	return RandomNumbers(6)
}

func (a Auth) AuthorizeSeller(userRepo interface {
	FindUserByID(id uint) (*domain.User, error)
}) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		authHeader := ctx.Get("Authorization")
		tokenUser, err := a.VerifyToken(authHeader)
		if err != nil {
			return ctx.Status(401).JSON(fiber.Map{
				"message": "authorization failed",
				"reason":  err.Error(),
			})
		}

		if tokenUser.ID == 0 {
			return ctx.Status(401).JSON(fiber.Map{
				"message": "authorization failed",
				"reason":  "invalid token",
			})
		}

		dbUser, err := userRepo.FindUserByID(tokenUser.ID)
		if err != nil {
			return ctx.Status(401).JSON(fiber.Map{
				"message": "authorization failed",
				"reason":  "user not found",
			})
		}

		if strings.ToLower(strings.TrimSpace(dbUser.UserType)) != domain.SELLER {
			return ctx.Status(401).JSON(fiber.Map{
				"message": "authorization failed",
				"reason":  "please join seller program to manage products",
			})
		}

		ctx.Locals("user", *dbUser)
		return ctx.Next()
	}
}
