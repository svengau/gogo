

# 🕺🏿💃 Gogo 🕺💃🏿

> a simple tool to run a command in a given environment.

you don't like seeing environment variables in your history or in ~/.shrc ? 😭

with production passwords and database urls in plain text ? 🙉 😱 💥

So here is gogo a simple tool to load vars from a `.gogo.yaml` file! 🎉🥳

gogo will first read in the `./.gogo.yaml` file (for project specific config), and if no file, it will try to read `~/.gogo.yaml` (for more global config).
And last but not least, gogo can encrypt your file using AES.
## Installation
### Using brew

```
  brew install svengau/tap/gogo
```

### From the source code

```
  git clone git@github.com:svengau/gogo.git
  make install
  make build
  ./gogo --version
```

### Manually

```
curl -Lo /usr/local/bin/gogo https://github.com/svengau/gogo/releases/download/v0.0.1/gogo-darwin-amd64
chmod +x /usr/local/bin/gogo
```

## Usage

### Command line

```
  NAME:
      gogo - a tool to run a command in a given environment

  USAGE:
    gogo <env> command [command arguments...]
    gogo [options] <env>

  OPTIONS:
      --list     list variables
      --encrypt  encrypt .gogo.yaml
      --decrypt  decrypt .gogo.yaml
      --version  display version
      --dry      dry mode (default: false)
      --verbose  verbose mode (default: false)
      --help     show help (default: false)
```

### Sample

With this sample config:

```
encrypted: false
envs:
  demo:
    MY_VAR: demo
    MY_VAR2: demo2
    TEST: sample
  production:
    MY_VAR: production
    MY_VAR2: production2
```

Some sample commands:

```
  gogo --verbose demo node -e "console.log('MY_VAR=' + process.env.MY_VAR)"
  # OUTPUT:
  # MY_VAR=demo

  gogo demo 'echo MY_VAR=$MY_VAR'
  # OUTPUT:
  # MY_VAR=demo

  gogo demo node -e "let i = 0; const inter = setInterval(()=>{console.log('MY_VAR=' + process.env.MY_VAR); i++; if (i > 3) clearInterval(inter)}, 500)"
  # STREAMLY OUTPUT:
  # MY_VAR=demo

```

## Alternative solutions

A couple of great solutions already exist, and may better fit your needs.

- [env-cmd](https://www.npmjs.com/package/env-cmd)
- [dotenv](https://www.npmjs.com/package/dotenv)
- [cross-env](https://www.npmjs.com/package/cross-env)
