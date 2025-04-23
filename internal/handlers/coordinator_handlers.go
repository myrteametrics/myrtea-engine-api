package handlers

// // GetAlias godoc
// // @Title GetAlias
// // @Description returns an alias based on an instance, a document type and a depth
// // @tags Alias
// // @Resource /coordinator
// // @Router /coordinator/alias [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func GetAlias(w http.ResponseWriter, r *http.Request) {
// 	instance := chi.URLParam(r, "instance")
// 	logicalIndex := fmt.Sprintf("%s-%s", instance, chi.URLParam(r, "documenttype"))
// 	depth := chi.URLParam(r, "depth")

// 	li, ok := coordinator.GetInstance().LogicalIndex(logicalIndex)
// 	if !ok {
// 		render.Error(w, r, render.ErrAPIDBResourceNotFound, fmt.Errorf("coordinator instance %s with index %s not found", instance, logicalIndex))
// 	}

// 	resp := make(map[string]string)
// 	resp["alias"] = fmt.Sprintf("%s-%s", li.Name, depth)

// 	render.JSON(w, r, resp)
// }

// // GetAllCron godoc
// // @Title GetAllCron
// // @Description returns an overview of every defined crons
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/crons [get]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func GetAllCron(w http.ResponseWriter, r *http.Request) {

// 	cronByIndice := make(map[string]interface{})
// 	for _, li := range coordinator.GetInstance().LogicalIndices {
// 		cronByIndice[li.Name] = map[string]interface{}{
// 			"name":     li.Name,
// 			"location": li.Cron.Location().String(),
// 			"exp":      li.Model.ElasticsearchOptions.Rollcron,
// 			"prev":     li.Cron.Entries()[0].Prev,
// 			"next":     li.Cron.Entries()[0].Next,
// 		}
// 	}

// 	render.JSON(w, r, cronByIndice)

// 	render.NotImplemented(w, r)
// }

// // StartAllCron godoc
// // @Title StartAllCron
// // @Description starts every defined crons
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/crons/start [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func StartAllCron(w http.ResponseWriter, r *http.Request) {
// 	for _, li := range coordinator.GetInstance().LogicalIndices {
// 		li.GetCron().Start()
// 	}

// 	render.OK(w, r)
// }

// // StopAllCron godoc
// // @Title StopAllCron
// // @Description stops every defined crons
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/crons/stop [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// // FIXME: Not working
// func StopAllCron(w http.ResponseWriter, r *http.Request) {
// 	for _, li := range coordinator.GetInstance().LogicalIndices {
// 		li.GetCron().Stop()
// 	}

// 	render.NotImplemented(w, r)
// }

// // GetCron godoc
// // @Title GetCron
// // @Description returns an overview of a single cron
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/cron [get]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// // TODO: ADD URL PARAMS SWAGGER ANNOTATION
// func GetCron(w http.ResponseWriter, r *http.Request) {
// 	logicalIndex := chi.URLParam(r, "logicalIndex")
// 	li := coordinator.GetInstance().LogicalIndex(logicalIndex)

// 	cron := make(map[string]interface{})
// 	cron["name"] = li.Name
// 	cron["location"] = li.Cron.Location().String()
// 	cron["exp"] = li.Model.ElasticsearchOptions.Rollcron
// 	cron["prev"] = li.Cron.Entries()[0].Prev
// 	cron["next"] = li.Cron.Entries()[0].Next

// 	render.JSON(w, r, cron)

// 	render.NotImplemented(w, r)
// }

// // StartCron godoc
// // @Title StartCron
// // @Description starts a single cron
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/cron/start [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func StartCron(w http.ResponseWriter, r *http.Request) {
// 	logicalIndex := chi.URLParam(r, "logicalIndex")

// 	coordinator.GetInstance().LogicalIndex(logicalIndex).GetCron().Start()

// 	render.NotImplemented(w, r)
// }

// // StopCron godoc
// // @Title StopCron
// // @Description stops a single cron
// // @tags Cron
// // @Resource /coordinator
// // @Router /coordinator/cron/stop [post]
// // @Accept json
// // @Success 200 "OK"
// // @Failure 400 "Bad Request"
// // @Failure 500 "Internal Server Error"
// func StopCron(w http.ResponseWriter, r *http.Request) {
// 	logicalIndex := chi.URLParam(r, "logicalIndex")

// 	coordinator.GetInstance().LogicalIndex(logicalIndex).GetCron().Stop()

// 	render.NotImplemented(w, r)
// }
