//go:build ignore

package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/echterhof/rockstar-web-framework/pkg"
)

func main() {
	// Configuration
	config := pkg.FrameworkConfig{
		ServerConfig: pkg.ServerConfig{
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			EnableHTTP1:  true,
			EnableHTTP2:  true,
		},
	}

	// Create framework instance
	app, err := pkg.New(config)
	if err != nil {
		log.Fatalf("Failed to create framework: %v", err)
	}

	// Create pipeline engine
	engine := pkg.NewPipelineEngine()

	// Pipeline 1: Data validation pipeline
	validationPipeline := pkg.PipelineConfig{
		Name: "validator",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Validate request data
			queryParams := ctx.Query()
			data, exists := queryParams["data"]
			if !exists || data == "" {
				return pkg.PipelineResultClose, fmt.Errorf("data parameter is required")
			}

			// Log validation success
			if ctx.Logger() != nil {
				ctx.Logger().Info("Data validated successfully")
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 100, // High priority - execute first
	}
	engine.Register(validationPipeline)

	// Pipeline 2: Data transformation pipeline
	transformPipeline := pkg.PipelineConfig{
		Name: "transformer",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Get data from query
			queryParams := ctx.Query()
			data, _ := queryParams["data"]

			// Transform data (uppercase)
			transformed := strings.ToUpper(data)

			// Log transformation
			if ctx.Logger() != nil {
				ctx.Logger().Info(fmt.Sprintf("Transformed: %s -> %s", data, transformed))
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 50,
	}
	engine.Register(transformPipeline)

	// Pipeline 3: Data enrichment pipeline
	enrichmentPipeline := pkg.PipelineConfig{
		Name: "enricher",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Get data from query
			queryParams := ctx.Query()
			data, _ := queryParams["data"]

			// Enrich with metadata
			enriched := map[string]interface{}{
				"data":      data,
				"length":    len(data),
				"timestamp": time.Now().Unix(),
				"processed": true,
			}

			// Log enrichment
			if ctx.Logger() != nil {
				ctx.Logger().Info(fmt.Sprintf("Enriched data: %v", enriched))
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 25,
	}
	engine.Register(enrichmentPipeline)

	// Pipeline 4: Slow processing pipeline with timeout
	slowPipeline := pkg.PipelineConfig{
		Name: "slow-processor",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Simulate slow processing
			queryParams := ctx.Query()
			delay, exists := queryParams["delay"]
			if exists && delay != "" {
				if ms, err := strconv.Atoi(delay); err == nil {
					time.Sleep(time.Duration(ms) * time.Millisecond)
				}
			}

			if ctx.Logger() != nil {
				ctx.Logger().Info("Slow processing complete")
			}
			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
		Timeout: 2000, // 2 second timeout
	}
	engine.Register(slowPipeline)

	// Pipeline 5: Async background task pipeline
	backgroundPipeline := pkg.PipelineConfig{
		Name: "background-task",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Simulate background processing
			time.Sleep(100 * time.Millisecond)

			// Log completion
			if ctx.Logger() != nil {
				ctx.Logger().Info("Background task completed")
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
		Async:   true,
	}
	engine.Register(backgroundPipeline)

	// Pipeline 6: Error handling pipeline
	errorHandlerPipeline := pkg.PipelineConfig{
		Name: "error-handler",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Check query for error trigger
			queryParams := ctx.Query()
			if triggerError, exists := queryParams["trigger_error"]; exists && triggerError == "true" {
				return pkg.PipelineResultClose, fmt.Errorf("simulated pipeline error")
			}

			return pkg.PipelineResultContinue, nil
		},
		Enabled:  true,
		Priority: 10,
	}
	engine.Register(errorHandlerPipeline)

	// Pipeline 7: Chaining pipeline
	chainStartPipeline := pkg.PipelineConfig{
		Name: "chain-start",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			if ctx.Logger() != nil {
				ctx.Logger().Info("Chain step: start")
			}
			return pkg.PipelineResultChain, nil
		},
		NextPipeline: "chain-middle",
		Enabled:      true,
	}
	engine.Register(chainStartPipeline)

	chainMiddlePipeline := pkg.PipelineConfig{
		Name: "chain-middle",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			if ctx.Logger() != nil {
				ctx.Logger().Info("Chain step: middle")
			}
			return pkg.PipelineResultChain, nil
		},
		NextPipeline: "chain-end",
		Enabled:      true,
	}
	engine.Register(chainMiddlePipeline)

	chainEndPipeline := pkg.PipelineConfig{
		Name: "chain-end",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			if ctx.Logger() != nil {
				ctx.Logger().Info("Chain step: end")
			}
			return pkg.PipelineResultContinue, nil
		},
		Enabled: true,
	}
	engine.Register(chainEndPipeline)

	// Pipeline 8: View rendering pipeline
	viewPipeline := pkg.PipelineConfig{
		Name: "view-renderer",
		Handler: func(ctx pkg.Context) (pkg.PipelineResult, error) {
			// Prepare data for view
			if ctx.Logger() != nil {
				ctx.Logger().Info("Preparing view data")
			}
			return pkg.PipelineResultView, nil
		},
		ViewHandler: func(ctx pkg.Context) error {
			data := map[string]interface{}{
				"title":   "Pipeline View",
				"message": "Rendered from pipeline",
			}
			return ctx.JSON(200, data)
		},
		Enabled: true,
	}
	engine.Register(viewPipeline)

	// Get router
	router := app.Router()

	// Route 1: Execute single pipeline
	router.GET("/pipeline/:name", func(ctx pkg.Context) error {
		name := ctx.Params()["name"]

		result, err := engine.Execute(ctx, name)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error":    err.Error(),
				"pipeline": name,
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"pipeline": name,
			"result":   result,
			"message":  "Pipeline executed successfully",
		})
	})

	// Route 2: Execute pipeline chain
	router.GET("/chain", func(ctx pkg.Context) error {
		// Execute chain of pipelines
		pipelines := []string{"validator", "transformer", "enricher"}

		result, err := engine.ExecuteChain(ctx, pipelines)
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error":     err.Error(),
				"pipelines": pipelines,
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"pipelines": pipelines,
			"result":    result,
			"message":   "Pipeline chain executed successfully",
		})
	})

	// Route 3: Execute async pipeline
	router.GET("/async", func(ctx pkg.Context) error {
		err := engine.ExecuteAsync(ctx, "background-task")
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"message": "Background task started",
			"status":  "running",
		})
	})

	// Route 4: Execute multiple pipelines concurrently
	router.GET("/multiplex", func(ctx pkg.Context) error {
		pipelines := []string{"validator", "transformer", "enricher"}

		results, errors := engine.ExecuteMultiplex(ctx, pipelines)

		// Collect results
		pipelineResults := make([]map[string]interface{}, len(pipelines))
		for i, name := range pipelines {
			pipelineResults[i] = map[string]interface{}{
				"pipeline": name,
				"result":   results[i],
			}
			if errors[i] != nil {
				pipelineResults[i]["error"] = errors[i].Error()
			}
		}

		return ctx.JSON(200, map[string]interface{}{
			"pipelines": pipelineResults,
		})
	})

	// Route 5: Test pipeline timeout
	router.GET("/timeout", func(ctx pkg.Context) error {
		result, err := engine.Execute(ctx, "slow-processor")

		if err != nil {
			return ctx.JSON(408, map[string]interface{}{
				"error":    err.Error(),
				"pipeline": "slow-processor",
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"pipeline": "slow-processor",
			"result":   result,
			"message":  "Slow processing completed",
		})
	})

	// Route 6: Test automatic chaining
	router.GET("/auto-chain", func(ctx pkg.Context) error {
		result, err := engine.Execute(ctx, "chain-start")
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"result":  result,
			"message": "Automatic chaining completed (check logs for chain steps)",
		})
	})

	// Route 7: Test view pipeline
	router.GET("/view", func(ctx pkg.Context) error {
		// Execute view pipeline (it will handle the response)
		_, err := engine.Execute(ctx, "view-renderer")
		return err
	})

	// Route 8: List all pipelines
	router.GET("/pipelines", func(ctx pkg.Context) error {
		pipelines := engine.List()

		pipelineInfo := make([]map[string]interface{}, len(pipelines))
		for i, pl := range pipelines {
			pipelineInfo[i] = map[string]interface{}{
				"name":     pl.Name,
				"enabled":  pl.Enabled,
				"priority": pl.Priority,
				"async":    pl.Async,
				"timeout":  pl.Timeout,
			}
		}

		return ctx.JSON(200, map[string]interface{}{
			"count":     len(pipelines),
			"pipelines": pipelineInfo,
		})
	})

	// Route 9: Enable/disable pipeline
	router.POST("/pipeline/:name/:action", func(ctx pkg.Context) error {
		name := ctx.Params()["name"]
		action := ctx.Params()["action"]

		var err error
		switch action {
		case "enable":
			err = engine.Enable(name)
		case "disable":
			err = engine.Disable(name)
		default:
			return ctx.JSON(400, map[string]interface{}{
				"error": "Invalid action. Use 'enable' or 'disable'",
			})
		}

		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error": err.Error(),
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"pipeline": name,
			"action":   action,
			"message":  fmt.Sprintf("Pipeline %s %sd", name, action),
		})
	})

	// Route 10: Error handling demonstration
	router.GET("/error-handling", func(ctx pkg.Context) error {
		result, err := engine.Execute(ctx, "error-handler")
		if err != nil {
			return ctx.JSON(500, map[string]interface{}{
				"error":   err.Error(),
				"handled": true,
			})
		}

		return ctx.JSON(200, map[string]interface{}{
			"result":  result,
			"message": "No errors detected",
		})
	})

	// Startup message
	fmt.Printf("🎸 Rockstar Web Framework - Pipeline Example\n")
	fmt.Printf("==============================================\n\n")
	fmt.Printf("Listening on :8080\n\n")
	fmt.Printf("Available endpoints:\n")
	fmt.Printf("  GET  /pipeline/:name                - Execute single pipeline\n")
	fmt.Printf("  GET  /chain                         - Execute pipeline chain\n")
	fmt.Printf("  GET  /async                         - Execute async pipeline\n")
	fmt.Printf("  GET  /multiplex                     - Execute multiple pipelines concurrently\n")
	fmt.Printf("  GET  /timeout                       - Test pipeline timeout\n")
	fmt.Printf("  GET  /auto-chain                    - Test automatic chaining\n")
	fmt.Printf("  GET  /view                          - Test view pipeline\n")
	fmt.Printf("  GET  /pipelines                     - List all pipelines\n")
	fmt.Printf("  POST /pipeline/:name/:action        - Enable/disable pipeline\n")
	fmt.Printf("  GET  /error-handling                - Test error handling\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  curl 'http://localhost:8080/pipeline/validator?data=test'\n")
	fmt.Printf("  curl 'http://localhost:8080/chain?data=hello'\n")
	fmt.Printf("  curl http://localhost:8080/async\n")
	fmt.Printf("  curl 'http://localhost:8080/multiplex?data=test'\n")
	fmt.Printf("  curl 'http://localhost:8080/timeout?delay=500'\n")
	fmt.Printf("  curl 'http://localhost:8080/timeout?delay=3000'  # Will timeout\n")
	fmt.Printf("  curl http://localhost:8080/auto-chain\n")
	fmt.Printf("  curl http://localhost:8080/view\n")
	fmt.Printf("  curl http://localhost:8080/pipelines\n")
	fmt.Printf("  curl -X POST http://localhost:8080/pipeline/validator/disable\n")
	fmt.Printf("  curl 'http://localhost:8080/error-handling?trigger_error=true'\n\n")
	fmt.Printf("Pipeline Features:\n")
	fmt.Printf("  - Sequential execution with priority ordering\n")
	fmt.Printf("  - Pipeline chaining (automatic and manual)\n")
	fmt.Printf("  - Async execution with goroutines\n")
	fmt.Printf("  - Concurrent execution (multiplexing)\n")
	fmt.Printf("  - Timeout support\n")
	fmt.Printf("  - Error handling and propagation\n")
	fmt.Printf("  - View rendering integration\n")
	fmt.Printf("  - Dynamic enable/disable\n\n")

	// Start server
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
