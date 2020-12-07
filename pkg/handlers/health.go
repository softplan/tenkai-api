package handlers

import (
	"net/http"

	"github.com/softplan/tenkai-api/pkg/global"
)

func (appContext *AppContext) healthRabbit(w http.ResponseWriter, r *http.Request) {
	_, err := appContext.RabbitMQChannel.QueueInspect("ResultInstallQueue")
	if err != nil {
		global.Logger.Error(global.AppFields{global.Function: "health"}, "Error when try to inspect queue")
		http.Error(w, "Error on RabbitMQ", http.StatusInternalServerError)
		return
	}
	http.Error(w, "", http.StatusOK)
}
