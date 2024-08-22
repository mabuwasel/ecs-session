
# ECS Session CLI Tool

🚀 **ECS Session** is an interactive CLI tool that allows you to connect to containers running on AWS ECS Fargate tasks. This tool provides an easy-to-use interface to select AWS regions, ECS clusters, services, and tasks, and then start an interactive session with the container of your choice using the AWS CLI `execute-command` feature.

## Prerequisites

Before using this tool, ensure you have the following installed on your system:

- **Go:** Install Go from [here](https://go.dev/dl/).
- **AWS CLI:** Ensure the AWS CLI is installed and configured with the necessary permissions to access ECS resources. Installation instructions can be found [here](https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html).
- **AWS IAM Permission** the Repository contains an IAM Policy JSON file template that you can use to create a new IAM policy and attach it to the IAM user or role that you use to access ECS resources. The policy grants the necessary permissions to list and describe ECS clusters, services, and tasks, as well as to execute commands on ECS tasks. You can find the policy template in the `iam-policy.json` file in the repository. you need to replace the following place holders in the policy : [ ```REGION,AWS_ACCOUNT_NUMBER,CLUSTER_NAME,SERVICE_NAME```] with your own values.
- **AWS CLI Session Manager Plugin:** The ECS Session tool uses the AWS CLI Session Manager plugin to establish an interactive session with the container. Ensure that the plugin is installed on your system. You can find installation instructions [here](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-install-plugin.html).

**NOTE** : The ECS service you select must have the execute-command feature enabled. This feature allows the tool to execute command and establish interactive sessios with the container. If you attempt to use this tool with a service that does not have execute-command enabled, the tool will detect this and log a message informing you of the issue.

## Installation

### Clone the Repository

Clone the repository to your local machine:

```bash
git clone https://github.com/mabuwasel/ecs-session.git
cd ecs-session
```

### Build the CLI Tool

Use the Go command to build the CLI tool:

```bash
go build -o ecs-session
```

### Basic Usage
Once the tool is built, you can start using it to connect to your ECS Fargate containers.

Start the Tool
To start the CLI tool, run:

```bash
./ecs-session
```

You can specify the AWS region directly using the --region or -r flag:

```bash
./ecs-session --region us-east-1
```

**Quick Demo**
![](https://raw.githubusercontent.com/mabuwasel/ecs-session/main/demo.gif)



