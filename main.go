package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/spf13/cobra"
)

const defaultRegionFile = "default_region.txt"

var region string

func main() {
	var rootCmd = &cobra.Command{
		Use:   "ecs-session",
		Short: "🚀 Interactive CLI tool for ECS Fargate task sessions",
		Run: func(cmd *cobra.Command, args []string) {
			startSession()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&region, "region", "r", "", "🌍 AWS Region (e.g., us-west-2)")
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func startSession() {
	// Check if a default region is stored in the local file
	if region == "" {
		region = loadDefaultRegion()
		if region != "" {
			fmt.Printf("ℹ️  Found saved region '%s'. Do you want to use it? (y/n): ", region)
			var useSaved string
			fmt.Scanf("%s", &useSaved)
			if strings.ToLower(useSaved) != "y" {
				region = ""
			}
		}
	}

	if region == "" {
		region = enterOrChooseRegion()
		saveRegionAsDefault(region)
	}

	clearScreen()
	fmt.Printf("✅ Region: %s\n", region)

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		log.Fatalf("❌ Unable to load SDK config: %v", err)
	}

	ecsClient := ecs.NewFromConfig(cfg)

	for {
		clusterArns, err := listClusters(ecsClient)
		if err != nil {
			log.Fatalf("❌ Unable to list clusters: %v", err)
		}

		clusterName := chooseOptionWithBack("cluster", clusterArns)
		if clusterName == "BACK" {
			region = ""
			break
		}
		clearScreen()
		fmt.Printf("✅ Region: %s\n", region)
		fmt.Printf("✅ Cluster: %s\n", clusterName)

		for {
			serviceArns, err := listServices(ecsClient, clusterName)
			if err != nil {
				log.Fatalf("❌ Unable to list services: %v", err)
			}

			serviceName := chooseOptionWithBack("service", serviceArns)
			if serviceName == "BACK" {
				break
			}

			// Check if the selected service has execute-command enabled
			describeOutput, err := ecsClient.DescribeServices(context.TODO(), &ecs.DescribeServicesInput{
				Cluster:  &clusterName,
				Services: []string{serviceName},
			})
			if err != nil {
				log.Fatalf("❌ Unable to describe services: %v", err)
			}

			service := describeOutput.Services[0]
			if !service.EnableExecuteCommand {
				clearScreen()
				fmt.Printf("⚠️  Execute-command is disabled for service: %s\n", serviceName)
				fmt.Println("Do you want to go back and choose a different service? (y/n): ")
				var goBack string
				fmt.Scanf("%s", &goBack)
				if strings.ToLower(goBack) == "y" {
					continue
				}
			}

			clearScreen()
			fmt.Printf("✅ Cluster: %s\n", clusterName)
			fmt.Printf("✅ Service: %s\n", serviceName)

			for {
				taskArns, err := listTasks(ecsClient, clusterName, serviceName)
				if err != nil {
					log.Fatalf("❌ Unable to list tasks: %v", err)
				}

				taskArn := chooseOptionWithBack("task", taskArns)
				if taskArn == "BACK" {
					break
				}
				clearScreen()
				fmt.Printf("✅ Cluster: %s\n", clusterName)
				fmt.Printf("✅ Service: %s\n", serviceName)
				fmt.Printf("✅ Task: %s\n", taskArn)

				for {
					containerNames, err := listContainers(ecsClient, clusterName, taskArn)
					if err != nil {
						log.Fatalf("❌ Unable to list containers: %v", err)
					}

					containerName := chooseOptionWithBack("container", containerNames)
					if containerName == "BACK" {
						break
					}
					clearScreen()
					fmt.Printf("✅ Cluster: %s\n", clusterName)
					fmt.Printf("✅ Service: %s\n", serviceName)
					fmt.Printf("✅ Task: %s\n", taskArn)
					fmt.Printf("✅ Container: %s\n", containerName)

					command := chooseCommand()
					clearScreen()
					fmt.Printf("✅ Cluster: %s\n", clusterName)
					fmt.Printf("✅ Service: %s\n", serviceName)
					fmt.Printf("✅ Task: %s\n", taskArn)
					fmt.Printf("✅ Container: %s\n", containerName)
					runAWSSession(clusterName, taskArn, containerName, command)

					// Session complete, exit or go back
					return
				}
			}
		}
	}
}

func enterOrChooseRegion() string {
	fmt.Println("🔍 Would you like to:")
	fmt.Println("1) Enter a region manually (e.g., us-west-2)")
	fmt.Println("2) Choose from the 5 most-used regions")

	var choice int
	fmt.Printf("➡️  Enter the number of your choice: ")
	fmt.Scanf("%d", &choice)

	if choice == 1 {
		var enteredRegion string
		fmt.Printf("➡️  Enter your desired region code: ")
		fmt.Scanf("%s", &enteredRegion)
		return enteredRegion
	} else {
		return chooseRegion()
	}
}

func chooseRegion() string {
	// Limiting to the 5 most-used regions
	topRegions := []string{
		"us-east-1",
		"us-west-2",
		"eu-west-1",
		"ap-southeast-1",
		"ap-northeast-1",
	}

	return chooseOption("region", topRegions)
}

func listClusters(client *ecs.Client) ([]string, error) {
	output, err := client.ListClusters(context.TODO(), &ecs.ListClustersInput{})
	if err != nil {
		return nil, err
	}

	return extractNamesFromArns(output.ClusterArns, "cluster"), nil
}

func listServices(client *ecs.Client, clusterArn string) ([]string, error) {
	output, err := client.ListServices(context.TODO(), &ecs.ListServicesInput{
		Cluster: &clusterArn,
	})
	if err != nil {
		return nil, err
	}

	return extractNamesFromArns(output.ServiceArns, "service"), nil
}

func listTasks(client *ecs.Client, clusterArn string, serviceArn string) ([]string, error) {
	output, err := client.ListTasks(context.TODO(), &ecs.ListTasksInput{
		Cluster:     &clusterArn,
		ServiceName: &serviceArn,
	})
	if err != nil {
		return nil, err
	}

	return output.TaskArns, nil
}

func listContainers(client *ecs.Client, clusterArn string, taskArn string) ([]string, error) {
	output, err := client.DescribeTasks(context.TODO(), &ecs.DescribeTasksInput{
		Cluster: &clusterArn,
		Tasks:   []string{taskArn},
	})
	if err != nil {
		return nil, err
	}

	var containerNames []string
	if len(output.Tasks) > 0 {
		for _, container := range output.Tasks[0].Containers {
			containerNames = append(containerNames, aws.ToString(container.Name))
		}
	}

	return containerNames, nil
}

func extractNamesFromArns(arns []string, resourceType string) []string {
	var names []string
	for _, arn := range arns {
		parts := strings.Split(arn, ":")
		if resourceType == "cluster" {
			names = append(names, strings.Split(parts[5], "/")[1]) // Extracting the cluster name
		} else if resourceType == "service" {
			names = append(names, strings.Split(parts[5], "/")[2]) // Extracting the service name
		} else {
			names = append(names, arn) // For tasks, keep the ARN intact
		}
	}
	return names
}

func runAWSSession(clusterArn string, taskArn string, containerName string, command string) {
	cmd := exec.Command("aws", "ecs", "execute-command",
		"--cluster", clusterArn,
		"--task", taskArn,
		"--container", containerName,
		"--interactive",
		"--command", command,
		"--region", region)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	fmt.Println("🚀 Starting AWS CLI execute-command session...")
	if err := cmd.Run(); err != nil {
		log.Printf("❌ Failed to start execute-command session: %v", err)
		if strings.Contains(err.Error(), "is not enabled") {
			log.Fatalf("❌ Service does not have execute-command enabled: %v", err)
		} else {
			log.Fatalf("❌ Failed to start execute-command session: %v", err)
		}
	}
}

func chooseCommand() string {
	fmt.Println("🔍 Choose a command to run:")
	fmt.Println("1) sh")
	fmt.Println("2) bash")
	fmt.Println("3) Enter custom command")

	var choice int
	fmt.Printf("➡️  Enter the number of your choice: ")
	fmt.Scanf("%d", &choice)

	switch choice {
	case 1:
		return "sh"
	case 2:
		return "bash"
	case 3:
		var customCommand string
		fmt.Printf("➡️  Enter your custom command: ")
		fmt.Scanf("%s", &customCommand)
		return customCommand
	default:
		fmt.Println("❌ Invalid choice, defaulting to 'sh'")
		return "sh"
	}
}

func chooseOption(entity string, options []string) string {
	fmt.Printf("🔍 Choose a %s:\n", entity)
	for i, option := range options {
		fmt.Printf("%s[%d]%s %s\n", yellow(), i+1, reset(), option)
	}

	var choice int
	fmt.Printf("➡️  Enter the number of your choice: ")
	fmt.Scanf("%d", &choice)

	return options[choice-1]
}

func chooseOptionWithBack(entity string, options []string) string {
	fmt.Printf("🔍 Choose a %s (or type '0' to go back):\n", entity)
	fmt.Printf("%s[0]%s Go back\n", yellow(), reset())

	for i, option := range options {
		fmt.Printf("%s[%d]%s %s\n", yellow(), i+1, reset(), option)
	}

	var choice int
	fmt.Printf("➡️  Enter the number of your choice: ")
	fmt.Scanf("%d", &choice)

	if choice == 0 {
		return "BACK"
	}
	return options[choice-1]
}

func yellow() string {
	return "\033[33m"
}

func reset() string {
	return "\033[0m"
}

// clearScreen clears the terminal screen
func clearScreen() {
	cmd := exec.Command("clear")
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Load the default region from a local file
func loadDefaultRegion() string {
	data, err := ioutil.ReadFile(defaultRegionFile)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("⚠️  Could not read default region file: %v", err)
		}
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Save the region to a local file as the default for next time
func saveRegionAsDefault(region string) {
	fmt.Printf("ℹ️  Would you like to save '%s' as the default region for next time? (y/n): ", region)
	var saveDefault string
	fmt.Scanf("%s", &saveDefault)

	if strings.ToLower(saveDefault) == "y" {
		err := ioutil.WriteFile(defaultRegionFile, []byte(region), 0644)
		if err != nil {
			log.Printf("⚠️  Could not save default region: %v", err)
		} else {
			fmt.Println("✅ Default region saved.")
		}
	}
}
