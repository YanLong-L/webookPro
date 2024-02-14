package main

import (
	"github.com/gin-gonic/gin"
	"webookpro/internal/events"
)

type App struct {
	web       *gin.Engine
	consumers []events.Consumer
}
