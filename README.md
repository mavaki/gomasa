# gomasa

The goal of our project was to create an anonymous email client that users could access via the web. We decided to use Go because it has goroutines, which makes it easy for our client to handle multiple requests coming in at once. Ideally, we would have handled all the data under a single goroutine, but for now we thought it was sufficient to use locks to prevent multiple users from accessing the data variables (recipient, subject, and body) at the same time. We then used a goroutine to actually send the email.
