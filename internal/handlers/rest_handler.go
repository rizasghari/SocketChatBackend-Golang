package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"socketChat/internal/enums"
	"socketChat/internal/errs"
	"socketChat/internal/models"
	"socketChat/internal/msgs"
	"socketChat/internal/services"
	"socketChat/internal/utils"
	"time"

	"strconv"

	"github.com/gin-gonic/gin"
)

type RestHandler struct {
	authService        *services.AuthenticationService
	chatService        *services.ChatService
	whiteboardService  *services.WhiteboardService
	fileManagerService *services.FileManagerService
}

func NewRestandler(
	authService *services.AuthenticationService,
	chatService *services.ChatService,
	whiteboardService *services.WhiteboardService,
	fileManagerService *services.FileManagerService,
) *RestHandler {
	return &RestHandler{
		authService:        authService,
		chatService:        chatService,
		whiteboardService:  whiteboardService,
		fileManagerService: fileManagerService,
	}
}

// Index godoc
// @Summary      Show home page
// @Description  Get home page html
// @Produce      html
// @Router       / [get]
func (rh *RestHandler) Index(ctx *gin.Context) {
	ctx.HTML(http.StatusOK, "index.html", nil)
}

// Login godoc
// @Summary      Login user to account
// @Description  get string by ID
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Account ID"
// @Success      200  {object}  models.Response
// @Failure      400  {object}  models.Response
// @Failure      404  {object}  models.Response
// @Failure      500  {object}  models.Response
// @Router       /login [post]
func (rh *RestHandler) Login(ctx *gin.Context) {
	var errors []error

	var loginData models.LoginRequestBody
	err := ctx.BindJSON(&loginData)
	if err != nil {
		log.Println("Error login data json binding:", err)
		errors = append(errors, errs.ErrInvalidRequestBody)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	loginResponse, loginErrs := rh.authService.Login(&loginData)
	if len(loginErrs) > 0 {
		errors = append(errors, loginErrs...)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    loginResponse,
	})

}

func (rh *RestHandler) Register(ctx *gin.Context) {
	var errors []error

	var user models.User
	err := ctx.BindJSON(&user)
	if err != nil {
		errors = append(errors, errs.ErrInvalidRequestBody)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	_, registerErrs := rh.authService.Register(&user)
	if len(registerErrs) > 0 {
		errors = append(errors, registerErrs...)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgUserCreatedSuccessfully,
	})
}

func (rh *RestHandler) CreateWhiteboard(ctx *gin.Context) {
	var errors []error
	var createWhiteboardRequest models.CreateWhiteboardRequest
	err := ctx.BindJSON(&createWhiteboardRequest)
	if err != nil {
		errors = append(errors, err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	creatorID := utils.GetUserIdFromContext(ctx)

	whiteboard := &models.Whiteboard{
		ConversationID: createWhiteboardRequest.ConversationID,
		Creator: creatorID,
	}

	// Check if the creator is member of the conversation
	isMember := rh.chatService.CheckUserInConversation(creatorID, createWhiteboardRequest.ConversationID)
	if (!isMember) {
		errors = append(errors, errs.ErrInvalidConversationId)
		ctx.AbortWithStatusJSON(http.StatusForbidden, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	whiteboard, err = rh.whiteboardService.CreateNewWhiteboard(whiteboard)
	if err != nil {
		errors = append(errors, err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    whiteboard,
	})
}

func (rh *RestHandler) CreateWhiteboardDrawn(ctx *gin.Context) {
	var errors []error
	var createDrawnRequest models.CreateDrawnRequest
	err := ctx.BindJSON(&createDrawnRequest)
	if err != nil {
		errors = append(errors, err)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	drawer := utils.GetUserIdFromContext(ctx)

	drawn := &models.Drawn{
		WhiteboardID: createDrawnRequest.WhiteboardId,
		Drawer: drawer,
	}

	// Check if the creator is member of the conversation
	isMember := rh.chatService.CheckUserInConversation(drawer, createDrawnRequest.ConversationID)
	if (!isMember) {
		errors = append(errors, errs.ErrInvalidConversationId)
		ctx.AbortWithStatusJSON(http.StatusForbidden, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	drawn, err = rh.whiteboardService.CreareNewDrawn(drawn)
	if err != nil {
		errors = append(errors, err)
		ctx.AbortWithStatusJSON(http.StatusNotFound, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    drawn,
	})
}

func (rh *RestHandler) CreateConversation(ctx *gin.Context) {
	var errors []error

	var createConversationRequestBody models.CreateConversationRequestBody
	err := ctx.BindJSON(&createConversationRequestBody)
	if err != nil {
		errors = append(errors, errs.ErrInvalidRequestBody)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	conversation, errors := rh.chatService.CreateConversation(&createConversationRequestBody)
	if len(errors) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    conversation,
	})
}

func (rh *RestHandler) GetAllUsersWithPagination(ctx *gin.Context) {
	page := ctx.Query("page")
	size := ctx.Query("size")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		sizeInt = 10
	}

	response, errs := rh.authService.GetAllUsersWithPagination(pageInt, sizeInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}
	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    response,
	})
}

func (rh *RestHandler) GetSingleUser(ctx *gin.Context) {
	id := ctx.Param("id")

	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidParams},
		})
		return
	}

	user, errs := rh.authService.GetSingleUser(idInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}
	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    user,
	})
}

func (rh *RestHandler) GetUserConversations(ctx *gin.Context) {
	id := ctx.Param("id")
	page := ctx.Query("page")
	size := ctx.Query("size")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		sizeInt = 10
	}

	idInt, err := strconv.Atoi(id)
	if err != nil || idInt < 1 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidParams},
		})
		return
	}

	conversationsResponse, errs := rh.chatService.GetUserConversations(uint(idInt), pageInt, sizeInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}
	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    conversationsResponse,
	})
}

func (rh *RestHandler) GetUserConversationsByToken(ctx *gin.Context) {
	page := ctx.Query("page")
	size := ctx.Query("size")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		sizeInt = 10
	}

	idInt := ctx.MustGet("user_id").(uint)
	if err != nil || idInt < 1 {
		log.Println("User id not found")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnauthorized},
		})
		return
	}

	conversationsResponse, errs := rh.chatService.GetUserConversations(idInt, pageInt, sizeInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}
	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    conversationsResponse,
	})
}

func (rh *RestHandler) UploadUserProfilePhoto(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	if userID < 1 {
		log.Println("User id not found")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnauthorized},
		})
		return
	}

	file, err := ctx.FormFile("profile_photo")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrNoFileUploaded},
		})
		return
	}

	src, err := file.Open()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnableToOpenUploadedFile},
		})
		return
	}
	defer src.Close()

	// Generate a unique file name based on user ID and original file extension
	fileExt := filepath.Ext(file.Filename)
	// to avoid client side caching after profile photo upload
	random := time.Now().Unix()
	fileName := fmt.Sprintf("user_profile_photo_%s_%d%s", strconv.Itoa(int(userID)), random, fileExt)

	// Upload the file to MinIO
	url, err := rh.fileManagerService.UploadUserProfilePhoto(fileName, src, file.Size, file.Header.Get("Content-Type"), enums.FILE_BUCKET_USER_PROFILE)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnableToUploadFile},
		})
		return
	}

	// Update the user profile photo URL in the database
	if updateErrs := rh.authService.UpdateUserProfilePhoto(userID, url); updateErrs != nil {
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnableToUpdateProfilePhoto},
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    url,
	})
}

func (rh *RestHandler) SaveMessage(ctx *gin.Context) {
	senderID := utils.GetUserIdFromContext(ctx)
	if senderID < 1 {
		log.Println("User id not found")
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrUnauthorized},
		})
		return
	}

	var messageRequest models.MessageRequest
	if err := ctx.ShouldBindJSON(&messageRequest); err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidRequest},
		})
		return
	}

	message := &models.Message{
		ConversationID: messageRequest.ConversationID,
		Content:        messageRequest.Content,
		SenderID:       senderID,
	}

	msg, errs := rh.chatService.SaveMessage(message)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}

	// Todo: Send socket event
	// Todo: Send notification

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    msg,
	})
}

func (rh *RestHandler) GetMessagesByConversationID(ctx *gin.Context) {
	conversationID := ctx.Param("id")

	page := ctx.Query("page")
	size := ctx.Query("size")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		sizeInt = 10
	}

	if conversationID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidRequest},
		})
		return
	}

	conversationIdUint, err := strconv.Atoi(conversationID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidRequest},
		})
		return
	}

	messages, errs := rh.chatService.GetMessagesByConversationId(uint(conversationIdUint), pageInt, sizeInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    messages,
	})
}

func (rh *RestHandler) UpdateUser(ctx *gin.Context) {
	var errors []error
	var updateUserRequest models.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&updateUserRequest); err != nil {
		errors = append(errors, errs.ErrInvalidRequest)
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	userID := utils.GetUserIdFromContext(ctx)
	if userID < 1 {
		errors = append(errors, errs.ErrUnauthorized)
		ctx.AbortWithStatusJSON(http.StatusUnauthorized, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errors,
		})
		return
	}

	updateUserRequest.ID = uint(userID)

	updatedUser, errs := rh.authService.UpdateUser(&updateUserRequest)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    updatedUser,
	})
}

func (rh *RestHandler) DiscoverUsers(ctx *gin.Context) {
	page := ctx.Query("page")
	size := ctx.Query("size")

	pageInt, err := strconv.Atoi(page)
	if err != nil || pageInt < 1 {
		pageInt = 1
	}

	sizeInt, err := strconv.Atoi(size)
	if err != nil || sizeInt < 1 {
		sizeInt = 10
	}

	userID := utils.GetUserIdFromContext(ctx)
	users, errs := rh.authService.GetNotContactedYetUsers(userID, pageInt, sizeInt)
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    users,
	})
}

func (rh *RestHandler) GetUserProfile(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	profile, errs := rh.authService.GetUserProfile(uint(userID))
	if len(errs) > 0 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  errs,
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    profile,
	})
}

func (rh *RestHandler) GetConversationUnReadMessagesForUser(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)
	conversationID := ctx.Param("id")

	if conversationID == "" {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidParams},
		})
		return
	}

	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil || conversationIDInt < 1 {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{errs.ErrInvalidParams},
		})
		return
	}

	count, err := rh.chatService.GetConversationUnReadMessagesForUser(uint(conversationIDInt), userID)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{err},
		})
		return
	}

	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    count,
	})
}

func (rh *RestHandler) GetUsersWhoHaveSentMessage(ctx *gin.Context) {
	concurrentParam := ctx.Param("concurrent")
	withMutexParam := ctx.Param("mutex")
	var concurrent = concurrentParam == "concurrent"
	var withMutex = withMutexParam == "mutex"
	users, err := rh.chatService.GetUsersWhoHaveSentMessage(concurrent, withMutex)
	if err != nil {
		ctx.AbortWithStatusJSON(http.StatusBadRequest, models.Response{
			Success: false,
			Message: msgs.MsgOperationFailed,
			Errors:  []error{err},
		})
		return
	}
	ctx.JSON(http.StatusOK, models.Response{
		Success: true,
		Message: msgs.MsgOperationSuccessful,
		Data:    users,
	})
}
