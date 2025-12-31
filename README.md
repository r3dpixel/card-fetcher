# Card Fetcher

A Go library for fetching character cards from various AI character platforms. Supports multiple sources and provides a
unified interface for retrieving character metadata and PNG character cards.

## Features

- **Multi-platform support**: Fetch character cards from 7+ platforms
- **Unified interface**: Consistent API across all supported sources
- **Metadata extraction**: Retrieve character information, creator details, and tags
- **Character card parsing**: Full support for PNG character card format
- **URL routing**: Automatically detect and route URLs to the correct fetcher
- **Integration testing**: Built-in tools to verify source integrations

## Supported Platforms

- Character Tavern
- ChubAI
- NyaiMe
- PepHop
- Pygmalion (requires credentials)
- WyvernChat
- JannyAI (requires cookies)

## Installation

```bash
go get github.com/r3dpixel/card-fetcher
```

## Quick Start

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/r3dpixel/card-fetcher/router"
)

func main() {
    // Create a router with environment-based configuration
    r := router.EnvConfigured()

    // Create a task from any character URL
    task, ok := r.TaskOf("https://chub.ai/characters/example/character-name")
    if !ok {
        fmt.Println("Invalid or unsupported URL")
        return
    }

    // Fetch both metadata and character card
    metadata, characterCard, err := task.FetchAll()
    if err != nil {
        fmt.Printf("Error fetching character: %v\n", err)
        return
    }

    fmt.Printf("Character: %s\n", metadata.Name)
    fmt.Printf("Creator: %s\n", metadata.Nickname)
    fmt.Printf("Tags: %v\n", metadata.Tags)
}
```

### Fetching Multiple Characters

```go
urls := []string{
    "https://chub.ai/characters/example/char1",
    "https://pygmalion.chat/characters/char2",
    "https://characterhub.org/characters/char3",
}

// Get tasks for multiple URLs
taskSlice := r.TaskSliceOf(urls...)

fmt.Printf("Valid URLs: %d\n", len(taskSlice.ValidURLs))
fmt.Printf("Invalid URLs: %d\n", len(taskSlice.InvalidURLs))

// Fetch all character cards
for _, task := range taskSlice.Tasks {
    metadata, card, err := task.FetchAll()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        continue
    }
    fmt.Printf("Fetched: %s\n", metadata.Name)
}
```

### Fetch Metadata Only

```go
task, ok := r.TaskOf("https://chub.ai/characters/example/character-name")
if !ok {
    return
}

// Fetch only metadata (lighter operation)
metadata, err := task.FetchMetadata()
if err != nil {
    fmt.Printf("Error: %v\n", err)
    return
}

fmt.Printf("Character: %s\n", metadata.Name)
fmt.Printf("Title: %s\n", metadata.Title)
fmt.Printf("Tagline: %s\n", metadata.Tagline)
fmt.Printf("Created: %v\n", metadata.CreateTime)
fmt.Printf("Updated: %v\n", metadata.UpdateTime)
```

### Custom Router Configuration

```go
import (
    "time"
    "github.com/r3dpixel/card-fetcher/router"
    "github.com/r3dpixel/card-fetcher/impl"
    "github.com/r3dpixel/toolkit/reqx"
)

// Create router with custom HTTP client settings
r := router.New(reqx.Options{
    RetryCount:        4,
    MinBackoff:        10 * time.Millisecond,
    MaxBackoff:        500 * time.Millisecond,
    DisableKeepAlives: true,
    Impersonation:     reqx.Chrome,
})

// Register specific platform builders
builders := impl.DefaultBuilders(impl.BuilderOptions{
    // Configure platform-specific options
})
r.RegisterBuilders(builders...)
```

## Configuration

Some platforms require authentication. Set the following environment variables:

```bash
# Pygmalion credentials
export PYGMALION_USERNAME="your_username"
export PYGMALION_PASSWORD="your_password"

# JannyAI cookies
export JANNY_CF_COOKIE="your_cloudflare_cookie"
export JANNY_USER_AGENT="your_user_agent"
```

Copy `env.example` to `.env` and fill in your credentials:

```bash
cp env.example .env
# Edit .env with your credentials
source .env
```

## Integration Testing

Check if a platform integration is working correctly:

```go
import "github.com/r3dpixel/card-fetcher/source"

status := r.CheckIntegration(source.ChubAI)
fmt.Printf("Integration status: %s\n", status)

// Possible statuses:
// - INTEGRATION SUCCESS
// - MISSING FETCHER
// - SOURCE DOWN
// - INVALID CREDENTIALS
// - MISSING REMOTE RESOURCE
// - MISMATCHED REMOTE RESOURCE
// - MISSING LOCAL RESOURCE
// - INTEGRATION FAILURE
```

## Advanced Usage

### Working with Tasks

```go
// Get tasks as a map (deduplicated by normalized URL)
taskBucket := r.TaskMapOf(urls...)

for normalizedURL, task := range taskBucket.Tasks {
    fmt.Printf("Processing: %s\n", normalizedURL)
    metadata, card, err := task.FetchAll()
    // ... handle result
}
```

### Accessing Router Information

```go
// List all registered sources
sources := r.Sources()
fmt.Printf("Available sources: %v\n", sources)

// Get all registered fetchers
fetchers := r.Fetchers()
for _, fetcher := range fetchers {
    fmt.Printf("Source: %s\n", fetcher.SourceID())
    fmt.Printf("Base URLs: %v\n", fetcher.BaseURLs())
}
```

## Error Handling

The library uses typed errors for better error handling:

```go
import "github.com/r3dpixel/card-fetcher/fetcher"

metadata, card, err := task.FetchAll()
if err != nil {
    errCode := fetcher.GetErrCode(err)
    switch errCode {
    case fetcher.InvalidCredentialsErr:
        fmt.Println("Invalid credentials - check your auth settings")
    case fetcher.FetchMetadataErr:
        fmt.Println("Failed to fetch metadata from source")
    case fetcher.MalformedMetadataErr:
        fmt.Println("Received invalid metadata from source")
    default:
        fmt.Printf("Error: %v\n", err)
    }
}
```

## Project Structure

```
card-fetcher/
├── fetcher/       # Core fetcher interfaces and utilities
├── impl/          # Platform-specific implementations
├── models/        # Data models (Metadata, CardInfo, etc.)
├── router/        # URL routing and task management
├── source/        # Source platform definitions
└── task/          # Task execution and workflow
```