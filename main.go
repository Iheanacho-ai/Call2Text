package main

import (
	"fmt"
    "log"
    "os"
	"net/http"

    "github.com/sendgrid/sendgrid-go"
    "github.com/sendgrid/sendgrid-go/helpers/mail"
	"github.com/joho/godotenv"

)

func HandleTwilioWebHooks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml")

	// Respond with TwiML to say a message and record
	response := `
		<Response>
			<Say>This call is being recorded.</Say>
			<Pause length="1"/>
			<Say>Please leave a voicemail after the beep.</Say>
			<Record 
				action="https://aa8b-102-89-34-223.ngrok-free.app/handle-transcription" 
				transcribe="true" 
				playBeep="true" 
				maxLength="120"
				transcribeCallback="https://aa8b-102-89-34-223.ngrok-free.app/handle-transcription"
			/>
		</Response>
	`
	fmt.Fprint(w, response)
}

func SendEmail(recordingURL, transcriptionText string) error {
	sendGridAPIKey := os.Getenv("SENDGRID_API_KEY")
	senderEmail := os.Getenv("SENDER_EMAIL")
	recipientEmail := os.Getenv("RECIPIENT_EMAIL")

	// Create the email content
	subject := "Voicemail Recording and Transcription"
	body := "You have a new voicemail recording:\n\nRecording URL: " + recordingURL + "\n\nTranscription: " + transcriptionText

	from := mail.NewEmail("Sender", senderEmail)
	to := mail.NewEmail("Recipient", recipientEmail)
	message := mail.NewSingleEmail(from, subject, to, body, body)

	// Send the email
	client := sendgrid.NewSendClient(sendGridAPIKey)
	response, err := client.Send(message)
	if err != nil {
		log.Printf("Error: %d", response.StatusCode)
		return err
	}

	return nil

}

func HandleTranscription(w http.ResponseWriter, r *http.Request) {
	// Retrieve the recording URL and transcription text from Twilio's request
	recordingURL := r.FormValue("RecordingUrl")
	transcriptionText := r.FormValue("TranscriptionText")

	if recordingURL == "" || transcriptionText == "" {
		http.Error(w, "Recording URL or transcription text not found", http.StatusBadRequest)
		return
	}


	// Send email with recording URL and transcription
	err := SendEmail(recordingURL, transcriptionText)
	if err != nil {
		http.Error(w, "Failed to send email", http.StatusInternalServerError)
		return
	}

	fmt.Println("Email sent successfully")
	w.WriteHeader(http.StatusOK)
}

func main() {

	err := godotenv.Load()
    if err != nil {
        log.Fatalf("Error loading .env file")
    }

	http.HandleFunc("/answer-call", HandleTwilioWebHooks)
	http.HandleFunc("/handle-transcription", HandleTranscription) 

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Failed to start server:", err)
	}
}
