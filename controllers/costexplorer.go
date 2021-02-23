package controllers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/costexplorer"
	"github.com/gin-gonic/gin"
	"gitlab.com/fibocloud/aws-billing/api_v2/form"
)

var credit = "Credit"
var refund = "Refund"

// ConstExplorerController struct
type ConstExplorerController struct {
	BaseController
}

// Init Controller
func (co ConstExplorerController) Init(router *gin.RouterGroup) {
	router.POST("/getcost", co.Get)       // GetCost
	router.POST("/forecast", co.Forecast) // Forecast
}

// Get cost
// @Summary Get cost
// @Description Show cost
// @Tags CostExporer
// @Accept json
// @Produce json
// @Param getCost body form.CostExplorerParams true "getCost"
// @Success 200 {object} structs.ResponseBody{body=costexplorer.GetCostAndUsageOutput}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /getcost [post]
func (co *ConstExplorerController) Get(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.CostExplorerParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}
	authUser := co.GetAuth(c)

	sess, sessError := co.DefaultSvc(authUser.Base.ID)
	if sessError != nil {
		co.SetError(http.StatusInternalServerError, sessError.Error())
		return
	}

	var groups []*costexplorer.GroupDefinition
	var eachGroup costexplorer.GroupDefinition

	if params.GroupName != "" {
		eachGroup.Type = aws.String("DIMENSION")
		eachGroup.Key = aws.String(params.GroupName)
		groups = append(groups, &eachGroup)
	}

	svc := costexplorer.New(sess)
	input := &costexplorer.GetCostAndUsageInput{
		Granularity: aws.String(params.Granularity),
		Metrics:     params.Metric,
		TimePeriod: &costexplorer.DateInterval{
			End:   aws.String(params.EndDate),
			Start: aws.String(params.StartDate),
		},
	}

	if len(groups) > 0 {
		input.GroupBy = groups
	}

	if len(params.Services) > 0 {
		serviceFilter := costexplorer.Expression{
			And: []*costexplorer.Expression{
				{
					Dimensions: &costexplorer.DimensionValues{
						Key:    aws.String("SERVICE"),
						Values: params.Services,
					},
				},
				{
					Not: &costexplorer.Expression{
						Dimensions: &costexplorer.DimensionValues{
							Key:    aws.String("RECORD_TYPE"),
							Values: []*string{&credit, &refund},
						},
					},
				},
			},
		}

		input.Filter = &serviceFilter
	} else {
		serviceFilter := costexplorer.Expression{
			Not: &costexplorer.Expression{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("RECORD_TYPE"),
					Values: []*string{&credit, &refund},
				},
			},
		}
		input.Filter = &serviceFilter
	}

	fmt.Println("input", input)

	cost, costErr := svc.GetCostAndUsage(input)
	if costErr != nil {
		co.SetError(http.StatusInternalServerError, costErr.Error())
		return
	}

	co.SetBody(cost)
	return
}

// Forecast cost
// @Summary Forecast cost
// @Description Forecast cost
// @Tags CostExporer
// @Accept json
// @Produce json
// @Param getForecastCost body form.CostExplorerForcastParams true "getForecastCost"
// @Success 200 {object} structs.ResponseBody{body=costexplorer.GetCostAndUsageOutput}
// @Failure 400 {object} structs.ErrorResponse
// @Failure 500 {object} structs.ErrorResponse
// @Router /forecast [post]
func (co *ConstExplorerController) Forecast(c *gin.Context) {
	defer func() {
		c.JSON(co.GetBody())
	}()

	var params form.CostExplorerForcastParams
	if err := c.ShouldBindJSON(&params); err != nil {
		co.SetError(http.StatusBadRequest, err.Error())
		return
	}
	authUser := co.GetAuth(c)

	sess, sessError := co.DefaultSvc(authUser.Base.ID)
	if sessError != nil {
		co.SetError(http.StatusInternalServerError, sessError.Error())
		return
	}

	startDate := time.Now().Add(time.Hour * 24).Format("2006-01-02")

	svc := costexplorer.New(sess)
	input := &costexplorer.GetCostForecastInput{
		Filter: &costexplorer.Expression{
			Not: &costexplorer.Expression{
				Dimensions: &costexplorer.DimensionValues{
					Key:    aws.String("RECORD_TYPE"),
					Values: []*string{&credit, &refund},
				},
			},
		},
		Granularity: aws.String(params.Granularity),
		Metric:      aws.String(params.Metric),
		TimePeriod: &costexplorer.DateInterval{
			End:   aws.String(params.EndDate),
			Start: aws.String(startDate),
		},
	}

	cost, errcost := svc.GetCostForecast(input)
	if errcost != nil {
		co.SetError(http.StatusInternalServerError, errcost.Error())
		return
	}

	co.SetBody(cost)
	return
}

// // CostUageWithResource ...
// // @Title CostUageWithResource
// // @Description CostUageWithResource
// // @Param    metric body    string    true    "metric"
// // @Param    granularity body    string    true    "granularity"
// // @Param    end_date body    date    true    "end_date"
// // @Failure 403
// // @router /costusagewithresource [post]
// func (e *ConstExplorerController) CostUageWithResource() {
// 	claims, _ := e.CheckToken()
// 	logger := utils.GetLogger()
// 	var response models.BaseResponse

// 	defer func() {
// 		e.Data["json"] = response
// 		e.ServeJSON()
// 	}()

// 	bodyResult := shared.RetrieveDataFromBody(e.Ctx.Request.Body)
// 	bodyString := []string{
// 		"metric",
// 		"granularity",
// 		"end_date",
// 		"start_date",
// 	}

// 	if sCod, eMsg := shared.CheckBodyResult(bodyResult, bodyString); sCod == 100 {
// 		response.StatusCode = sCod
// 		response.ErrorMsg = eMsg
// 		return
// 	}

// 	userEmail := claims["email"].(string)

// 	metrics := bodyResult["metric"].([]interface{})
// 	// granularity := bodyResult["granularity"].(string)
// 	startDate := bodyResult["start_date"].(string)
// 	endDate := bodyResult["end_date"].(string)

// 	var metricsValue []*string

// 	for _, i := range metrics {
// 		eachID := i.(string)
// 		metricsValue = append(metricsValue, &eachID)
// 	}

// 	sess, sessError := DefaultSvc(userEmail)
// 	if sessError != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(userEmail)
// 		logger.Error(sessError.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = sessError.Error()
// 		return
// 	}

// 	svc := costexplorer.New(sess)

// 	input := &costexplorer.GetCostAndUsageWithResourcesInput{
// 		Filter: &costexplorer.Expression{
// 			Dimensions: &costexplorer.DimensionValues{
// 				Key: aws.String("SERVICE"),
// 				Values: []*string{
// 					aws.String("AWS Amplify"),
// 					aws.String("AWS AppSync"),
// 					aws.String("AWS Data Transfer"),
// 					aws.String("AWS Elemental MediaStore"),
// 					aws.String("AWS Key Management Service"),
// 					aws.String("AWS Lambda"),
// 					aws.String("AWS Secrets Manager"),
// 					aws.String("Amazon API Gateway"),
// 					aws.String("Amazon CloudFront"),
// 					aws.String("Amazon Cognito"),
// 					aws.String("Amazon DynamoDB"),
// 					aws.String("Amazon EC2 Container Registry (ECR)"),
// 					aws.String("EC2 - Other"),
// 					aws.String("Amazon Elastic Compute Cloud - Compute"),
// 					aws.String("Amazon Elasticsearch Service"),
// 					aws.String("Amazon Glacier"),
// 					aws.String("Amazon Lex"),
// 					aws.String("Amazon Lightsail"),
// 					aws.String("Amazon Rekognition"),
// 					aws.String("Amazon Relational Database Service"),
// 					aws.String("Amazon Route 53"),
// 					aws.String("Amazon Simple Email Service"),
// 					aws.String("Amazon Simple Notification Service"),
// 					aws.String("Amazon Simple Queue Service"),
// 					aws.String("Amazon Simple Storage Service"),
// 					aws.String("AmazonCloudWatch"),
// 				},
// 			},
// 		},
// 		GroupBy: []*costexplorer.GroupDefinition{
// 			{
// 				Key:  aws.String("RESOURCE_ID"),
// 				Type: aws.String(costexplorer.GroupDefinitionTypeDimension),
// 			},
// 		},
// 		Granularity: aws.String(costexplorer.GranularityMonthly),
// 		Metrics:     metricsValue,
// 		TimePeriod: &costexplorer.DateInterval{
// 			End:   aws.String(endDate),
// 			Start: aws.String(startDate),
// 		},
// 	}

// 	cost, err := svc.GetCostAndUsageWithResources(input)

// 	if err != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(claims["email"].(string))
// 		logger.Error(err.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = err.Error()
// 		response.Body = true
// 		return
// 	}

// 	response.StatusCode = 0
// 	response.ErrorMsg = ""
// 	response.Body = cost
// 	return
// }

// // GetDimensionValues ...
// // @Title GetDimensionValues
// // @Description GetDimensionValues
// // @Param    start_date body    string    true    "start_date"
// // @Param    end_date body    date    true    "end_date"
// // @Failure 403
// // @router /getdimensionvalues [post]
// func (e *ConstExplorerController) GetDimensionValues() {
// 	claims, _ := e.CheckToken()
// 	logger := utils.GetLogger()
// 	var response models.BaseResponse

// 	defer func() {
// 		e.Data["json"] = response
// 		e.ServeJSON()
// 	}()

// 	bodyResult := shared.RetrieveDataFromBody(e.Ctx.Request.Body)
// 	bodyString := []string{
// 		"start_date",
// 		"end_date",
// 	}

// 	if sCod, eMsg := shared.CheckBodyResult(bodyResult, bodyString); sCod == 100 {
// 		response.StatusCode = sCod
// 		response.ErrorMsg = eMsg
// 		return
// 	}

// 	userEmail := claims["email"].(string)

// 	startDate := bodyResult["start_date"].(string)
// 	endDate := bodyResult["end_date"].(string)

// 	sess, sessError := DefaultSvc(userEmail)
// 	if sessError != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(userEmail)
// 		logger.Error(sessError.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = sessError.Error()
// 		return
// 	}

// 	svc := costexplorer.New(sess)
// 	input := &costexplorer.GetDimensionValuesInput{
// 		Context:   aws.String("COST_AND_USAGE"),
// 		Dimension: aws.String("SERVICE"),
// 		TimePeriod: &costexplorer.DateInterval{
// 			End:   aws.String(endDate),
// 			Start: aws.String(startDate),
// 		},
// 	}
// 	cost, err := svc.GetDimensionValues(input)

// 	if err != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(claims["email"].(string))
// 		logger.Error(err.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = err.Error()
// 		response.Body = true
// 		return
// 	}

// 	response.StatusCode = 0
// 	response.ErrorMsg = ""
// 	response.Body = cost
// 	return
// }

// // DescribeReportDefinitionsInput ...
// // @Title DescribeReportDefinitionsInput
// // @Description DescribeReportDefinitionsInput
// // @Param    start_date body    string    true    "start_date"
// // @Param    end_date body    date    true    "end_date"
// // @Failure 403
// // @router /describereportdefinitionsinput [post]
// func (e *ConstExplorerController) DescribeReportDefinitionsInput() {
// 	claims, _ := e.CheckToken()
// 	logger := utils.GetLogger()
// 	var response models.BaseResponse

// 	defer func() {
// 		e.Data["json"] = response
// 		e.ServeJSON()
// 	}()

// 	// bodyResult := shared.RetrieveDataFromBody(e.Ctx.Request.Body)
// 	// bodyString := []string{
// 	// 	"start_date",
// 	// 	"end_date",
// 	// }

// 	// if sCod, eMsg := shared.CheckBodyResult(bodyResult, bodyString); sCod == 100 {
// 	// 	response.StatusCode = sCod
// 	// 	response.ErrorMsg = eMsg
// 	// 	return
// 	// }

// 	userEmail := claims["email"].(string)

// 	sess, sessError := DefaultSvc(userEmail)
// 	if sessError != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(userEmail)
// 		logger.Error(sessError.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = sessError.Error()
// 		return
// 	}

// 	svc := costandusagereportservice.New(sess)
// 	input := &costandusagereportservice.DescribeReportDefinitionsInput{
// 		MaxResults: aws.Int64(5),
// 	}

// 	result, err := svc.DescribeReportDefinitions(input)
// 	if err != nil {
// 		if aerr, ok := err.(awserr.Error); ok {
// 			switch aerr.Code() {
// 			case costandusagereportservice.ErrCodeInternalErrorException:
// 				fmt.Println(costandusagereportservice.ErrCodeInternalErrorException, aerr.Error())
// 			default:
// 				fmt.Println(aerr.Error())
// 			}
// 		} else {
// 			// Print the error, cast err to awserr.Error to get the Code and
// 			// Message from an error.
// 			fmt.Println(err.Error())
// 		}
// 		return
// 	}

// 	if err != nil {
// 		e.RWMutex.Lock()
// 		defer e.RWMutex.Unlock()
// 		utils.SetLumlog(claims["email"].(string))
// 		logger.Error(err.Error())
// 		response.StatusCode = 100
// 		response.ErrorMsg = err.Error()
// 		response.Body = true
// 		return
// 	}

// 	response.StatusCode = 0
// 	response.ErrorMsg = ""
// 	response.Body = result
// 	return
// }
