# mailer

Simple transactional mailer middleware. Listens for POST HTTP requests using
webserver, sends emails using provider API upon request.

To send an email you need to send POST request to `/send` endpoint with data
being form-encoded (header `application/x-www-form-urlencoded`).

### Features

* Multiple transactional mail providers
    * Mandrill
    * Mailgun
    * TODO: SpartPost
    * TODO: SendGrid

* [MJML] templates support (if `mjml` app is available)

* Simplest API possible

* Async send


### API parameters

* `token` - mandatory, auth token registered in application's configuration file
* `subject` - email's subject
* `from_email` - mandatory, sender's email address
* `from_name` - sender's name
* `to` - recepient's email address, only one address is currently supported
* `text` - message in text format
* `html` - message in html format
* `mjml` - message in [mjml](http://mjml.io) format, it will be converted to html on fly

*Note*, that if `html` parameter is provided, then `mjml` parameter is ignored.
If both `text` and `html` (or `mjml`) parameters are provided then resulting
email will have both html and text parts (what part will be visible by user
depends on user mail client software).


### Return codes

Application returns result in JSON encoded format.

Structure of response:

    {"Status": "<status>", "Message": "<message>"}

* `Status` - can be either `ok`, or `error`. If `error` then corresponding
  non-200 HTTP return code will be returned
* `Message` - string, explaining status message


### Requirements

1. Go runtime
2. [MJML] when using mjml templates

### Example

1. Download and build

`go get github.com/schmooser/go-mailer`

2. Edit `config.toml` file and set up transactional mail provider

3. Run `mailer` app

4. Send test email

    curl -v -d "token=78adfh3jjduu&subject=this is subject with spaces&from_email=schmooser@gmail.com&to=darth.vader@mailinator.com&text=Hi, my lord!" \
    http://localhost:1233/send


[MJML]: http://mjml.io
