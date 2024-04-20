package utils

const (
	NOTION_DATABASE_ID = "NOTION_DATABASE_ID"
	TELEGRAM_TOKEN     = "TELEGRAM_TOKEN"
	DATABASE_URL       = "DATABASE_URL"
	OAUTH_URL          = "OAUTH_URL"
	PORT               = "PORT"
	VAULT_PATH         = "VAULT_PATH"
	VAULT_ADDR         = "VAULT_ADDR"
	VAULT_SECRET_KEY   = "VAULT_SECRET_KEY"
	VAULT_TOKEN        = "VAULT_TOKEN"

	MAX_LEN_MESSAGE = 4096
)

const (
	HELP_STRING = `Hi there 👋 I'm your personal bridge to Notion, designed to make noting down your thoughts, tasks, and ideas as easy as sending a message to a friend. Let's get your productivity supercharged without ever leaving Telegram
Here is how to get started:
	
- /help - Displays this help message;
- /register - Register your Notion notebook in the bot;
- /note text - Write the text of the note or upload a pdf, jpg, jpeg or png (if it's an image, please ensure to send it without compression or it will upload a blurred image on notion) on Notion;
- /defaultpage page_name - Sets the default Notion page for your notes. Ensure this is an authorized page during registration;
- /getdefaultpage - Get default page for your user;
- /deauthorize - I will forget you;
	
Need a bit more guidance? Type /help anytime to see what I can do for you or look at the Github repository: https://github.com/fulviodenza/notion-echo
	
Remember, your privacy is paramount. I don't keep any of your data. Everything goes straight into your Notion, and nowhere else.`
)

const (
	COMMAND_NOTE             = "/note"
	COMMAND_HELP             = "/help"
	COMMAND_REGISTER         = "/register"
	COMMAND_START            = "/start"
	COMMAND_DEFAULT_PAGE     = "/defaultpage"
	COMMAND_DEAUTHORIZE      = "/deauthorize"
	COMMAND_GET_DEFAULT_PAGE = "/getdefaultpage"
)

const (
	PRIVATE_CHAT_TYPE    = "private"
	GROUP_CHAT_TYPE      = "group"
	SUPERGROUP_CHAT_TYPE = "supergroup"
)
