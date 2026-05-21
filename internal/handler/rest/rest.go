package rest

import (
	"fmt"
	"greentrust-hackathon/internal/service"
	"greentrust-hackathon/pkg/middleware"
	"os"

	"github.com/gin-gonic/gin"
)

type Rest struct {
	router     *gin.Engine
	service    *service.Service
	middleware middleware.Interface
}

func NewRest(service *service.Service, middleware middleware.Interface) *Rest {
	return &Rest{
		router:     gin.Default(),
		service:    service,
		middleware: middleware,
	}
}

func (r *Rest) MountEndpoint() {
	r.router.Use(r.middleware.Cors())
	baseURL := r.router.Group("/api/v1")

	auth := baseURL.Group("/auth")
	auth.POST("/register", r.RegisterUser)
	auth.POST("/verify-otp", r.VerifyOTP)
	auth.POST("/login", r.LoginUser)

	baseURL.GET("/sectors", r.GetBusinessSectors)

	onboarding := baseURL.Group("/onboarding")
	onboarding.POST("/identity", r.SubmitUserIdentity)
	onboarding.POST("/business-profile", r.SubmitBusinessProfile)
	onboarding.POST("/investor/positions", r.CreateOnboardingInvestorPosition)
	onboarding.GET("/investor/positions", r.GetOnboardingInvestorPositions)
	onboarding.PUT("/investor/positions/:position_id", r.UpdateOnboardingInvestorPosition)
	onboarding.DELETE("/investor/positions/:position_id", r.DeleteOnboardingInvestorPosition)
	onboarding.GET("/investor/skills", r.SearchSkills)

	investor := baseURL.Group("/investor")
	investor.Use(r.middleware.AuthenticateUser)
	investor.POST("/positions", r.CreateInvestorPosition)
	investor.GET("/positions", r.GetInvestorPositions)
	investor.PUT("/positions/:position_id", r.UpdateInvestorPosition)
	investor.DELETE("/positions/:position_id", r.DeleteInvestorPosition)
	investor.GET("/skills", r.SearchSkills)

	evidence := baseURL.Group("/evidence")
	evidence.Use(r.middleware.AuthenticateUser)
	evidence.GET("/categories", r.GetEvidenceCategories)
	evidence.GET("/summary", r.GetEvidenceSummary)
	evidence.GET("/ai-reviews", r.GetEvidenceAIReviews)
	evidence.GET("/categories/:category_id", r.GetEvidenceCategoryDetail)
	evidence.POST("/categories/:category_id/documents", r.UploadEvidenceDocument)
	evidence.PATCH("/documents/:evidence_id/ai-review", r.SubmitEvidenceAIReview)
	evidence.PATCH("/ai-reviews/:review_id", r.ReviewEvidenceAI)

	greenPassport := baseURL.Group("/green-passports")
	greenPassport.Use(r.middleware.AuthenticateUser)
	greenPassport.POST("/issue", r.IssueGreenPassport)

	proposals := baseURL.Group("/proposals")
	proposals.Use(r.middleware.AuthenticateUser)
	proposals.POST("", r.CreateProposal)
	proposals.GET("", r.GetProposals)
	proposals.GET("/:proposal_id", r.GetProposalDetail)
	proposals.PUT("/:proposal_id", r.UpdateProposal)
	proposals.POST("/:proposal_id/send", r.SendProposal)
	proposals.POST("/:proposal_id/accept", r.AcceptProposal)
	proposals.POST("/:proposal_id/reject", r.RejectProposal)
	proposals.POST("/:proposal_id/withdraw", r.WithdrawProposal)

}

func (r *Rest) Run() {
	addr := os.Getenv("ADDRESS")
	port := os.Getenv("PORT")

	r.router.Run(fmt.Sprintf("%s:%s", addr, port))
}
