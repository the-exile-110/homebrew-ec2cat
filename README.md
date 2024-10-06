# ec2cat

ec2cat is a command-line tool for managing and monitoring AWS EC2 instances across multiple regions. It provides a user-friendly interface to view instance details, estimated costs, and running times.

## Features

- List EC2 instances across all AWS regions or a specific region
- Display instance details including name, ID, type, state, and launch time
- Calculate estimated hourly and total costs for running instances
- Show total runtime for each instance
- Support for multiple AWS profiles

## Installation

To install ec2cat, you can use Homebrew:

```bash
brew tap the-exile-110/ec2cat
brew install ec2cat
```

## Usage

1. Run the ec2cat command:
   ```bash
   ec2cat
   ```
2. Select your AWS profile when prompted.
3. Choose to view instances from all regions or select a specific region.
4. View the table of EC2 instances with their details and costs.

## Requirements

- Go 1.22 or higher
- AWS CLI configured with valid credentials

## Building from Source

1. Clone the repository:
   ```bash
   git clone https://github.com/the-exile-110/homebrew-ec2cat.git
   cd homebrew-ec2cat
   ```
2. Build the project:
   ```bash
   go build
   ```
3. Run the compiled binary:
   ```bash
   ./ec2cat
   ```

## Configuration

ec2cat uses the AWS CLI configuration. Ensure you have valid AWS credentials set up in your `~/.aws/credentials` file.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

MIT

## Acknowledgements

This project uses the following open-source libraries:

- [github.com/aws/aws-sdk-go-v2](https://github.com/aws/aws-sdk-go-v2)
- [github.com/manifoldco/promptui](https://github.com/manifoldco/promptui)
- [github.com/olekukonko/tablewriter](https://github.com/olekukonko/tablewriter)

For a complete list of dependencies, please refer to the `go.mod` file.

