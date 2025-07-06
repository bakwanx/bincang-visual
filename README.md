# Bincang Visual Go

Bincang visual is a web app meeting platform based on Flutter and Go. it allows users to make meeting within a minute without cost, fees and no login required. Just create new meeting and ready to go!

This project is still under development.
But still, you can try for live demo [here](https://bakwanx.github.io/bincang-visual-web/)

## Setup Instructions

To use this project, follow this steps:

1. **Clone the repository**: Clone this repository to your machine
2. **Install Golang**: Ensure you have installed Golang on your machine
3. **Setup**: .env
4. **Install Dependencies**: Navigate to the project directory and install the required dependencies
5. **Setup Redis**: Setup redis and add new key "config:coturn" and fill the value with this format

```
{
    "iceServers": [
        {
            "urls": [
                "turn:your-turn-server"
            ],
            "username": "",
            "credential": ""
        },
        {
            "urls": [
                "stun:your-stun-server",
                "stun:stun.flashdance.cx:3478" => example public STUN
            ]
        }
    ]
}
```

6. **Run the Service:** Now, you can run the service
