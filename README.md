# TexBot

Discord bot that renders server messages and DMs containing LaTeX expressions and sends them as images in chat. Once running, the bot will pick up any messages containing the following syntax:

`$$ <LaTeX math expression> $$`

You can also pass an *optional* integer size modifier to the expression (default size is `4`):

`$$ <LaTeX math expression> $$[<size>]`

The following examples are valid messages that will be parsed and renderd by TexBot:

```latex
% Render basic expression
$$ \frac{1}{ \sqrt{2} } $$

% Render text expression with LaTeX logo and a size modifier
$$ \text{ {\LaTeX} is awesome! } $$[3]

% Render multi-line expression using pmatrix command and a size modifier
% This command can also be written on a single line at the expense of readability
$$ \begin{pmatrix}
      0 & 1 \\
      1 & 0
   \end{pmatrix} $$[1]
```

### Prerequisites

* Discord account with [Discord API application](https://discord.com/developers/applications)
* [Discord bot user](https://discord.com/developers/docs/topics/oauth2#bots) for developer application
  * At the very least, the bot must have the following [permissions](https://discord.com/developers/docs/topics/oauth2#shared-resources-oauth2-scopes):
    * Read Messages
    * Send Messages
    * Send Messages in Threads
    * Manage Messages
    * Manage Threads
    * Embed Links
    * Attach Files
  * The permission bitstring `292057836544` has all of these basic permissions enabled
  * The OAuth2 URL should look something like this:
    * `https://discord.com/oauth2/authorize?client_id=XXXXX&permissions=292057836544&scope=bot`
  * Consult the Discord API documentation for more information about creating bots and OAuth2 permission URLs

### Running TexBot

TexBot lives entirely in a Docker container, and can be built and run using the following commands:

```bash
# Build the Docker image
# (this will take a while; total size is ~4GB for LaTeX libraries + fonts)
docker build -t texbot .

# Run the Docker container with Discord Bot token:
docker run texbot -t "<API token>"

# Alternatively, set the token as environment variable:
env DISCORD_TOKEN="<API token>" docker run texbot
```

TexBot is now running, simply invite it to a Discord server using the OAuth2 URL described above, and you're ready to go!

### Project Files

* `texbot.go` -- main executable and entrypoint for Docker container. Registers bot with Discord and waits for messages
* `parser.go` -- parses raw Discord messages and returns a successful/unsuccessful parse object containing a cleaned expression
  * `parser_test.go` -- file to test basic parsing functionality (such as cleaning input and setting size)
* `converter.go` -- converts LaTeX expressions to images in a temporary directory in the Docker container and returns an object containing a pointer to an in-memory image

---

## Original Proposal

* For my project, I would like to make a Discord bot written in Golang that watches for messages in the chat containing LaTeX expressions, renders them as images, and sends it in the chat

### Features/Background

* The basics of a Discord Bot is that it is registered with the Discord Developer API and attached to either one or many servers
  * There are permissions that are associated with the bot that are used to restrict the bot to only doing certain actions (e.g. reading chat messages, sending messages, etc.)
  * The bot can be added to a server by anyone with the ability to add bots to a server (typically the server owner)
* TexBot will run on Warloc and will listen for messages that contain LaTeX expressions (or will receive a message that begins with the command `!tex`)
  * TexBot will parse the LaTeX expression and render it to an image (likely on a different thread)
  * TexBot will then send the image to the channel that the message was sent in, and will delete the original message
    * Alternatively, if possible, it will edit the original message and attach the image so it appears to come from the same user as the original message
* The intention is also to run TexBot in a Docker container
  * The general build/run process may be like the following:

    ```shell
    docker build . -t texbot
    docker run texbot --token <bot API token>
    ```

### Resources

* I have located two Go modules that I will utilize for this project:
  1. [DiscordGo](https://github.com/bwmarrin/discordgo) -- a Go module to interact with the Discord API
  2. [go-latex](https://github.com/go-latex/latex) -- a Go module to render LaTeX expressions as images
     * After initial inspection, this may not be a perfect library to use, as it seems to implement its own rudimentary LaTeX parser
     * Another option is to try to use [this Node.js project](https://github.com/joeraut/latex2image-web) as a server that the bot will make web requests to using an internal Docker-Compose network
     * A final resort may be to use a command-like LaTeX parser/renderer, though I'd like to avoid this if possible

### Basic Flow

1. Save LaTeX source in `source.tex`
2. Run `latex` on `source.tex` to generate `source.dvi`
3. Run `dvisvgm` on `source.dvi` to generate `source.svg`
4. Convert `source.svg` to `source.png`
5. Send image
