# tts
`tts` is a text-to-speech CLI using [Google Cloud Text-to-Speech](https://cloud.google.com/text-to-speech).

## Usage
```bash
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials/file.json
echo 'Hello, world!' | tts --player mpv
```
