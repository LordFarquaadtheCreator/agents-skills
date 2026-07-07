---
name: deviantart
description: Browse, search, and interact with DeviantArt via the deviantart-mcp MCP server. Use when Fahad wants to find art, browse DeviantArt, check messages, manage collections, or view galleries.
---

# DeviantArt

Browse and interact with DeviantArt via the `deviantart-mcp` MCP server. OAuth-based — public browsing works with client credentials, user actions (fave, collections) require authorization.

## Tools by group

### Browse (public)
| Tool | What it does |
|---|---|
| `browse_popular` | Popular deviations |
| `browse_newest` | Newest deviations |
| `browse_hot_topics` | Current hot topics |
| `browse_tags` | Browse by tag |
| `browse_topic` | Specific topic |
| `browse_dailydeviations` | Daily deviations |
| `browse_journals` | Featured journals |
| `browse_more_like_this` | Related deviations |
| `browse_category_tree` | Category tree |
| `browse_user_friends` | User's friends |
| `browse_user_journals` | User's journals |
| `browse_user_literature` | User's literature |

### Deviation (single artwork)
| Tool | What it does |
|---|---|
| `deviation_get` | Fetch deviation by ID |
| `deviation_content` | Get content |
| `deviation_download` | Get download link |
| `deviation_metadata` | Get metadata |
| `deviation_whofaved` | Who faved it |
| `deviation_embedded_content` | Embedded content |

### Gallery (user's gallery)
| Tool | What it does |
|---|---|
| `gallery_all` | All deviations by user |
| `gallery_folders` | Gallery folders |
| `gallery_folder` | Specific folder |
| `gallery_folder_create` | Create folder |
| `gallery_folder_delete` | Delete folder |

### Collections (favorites)
| Tool | What it does |
|---|---|
| `collections_folders` | Collection folders |
| `collections_folder` | Specific collection |
| `collections_create` | Create collection |
| `collections_delete` | Delete collection |
| `collections_fave` | Favorite a deviation |
| `collections_unfave` | Unfavorite |

### Messages (notifications)
| Tool | What it does |
|---|---|
| `messages_feedback` | Feedback messages |
| `messages_feedback_stack` | Specific feedback stack |
| `messages_mentions` | Mentions |
| `messages_mentions_stack` | Specific mentions stack |

## Auth

- Public browsing (browse_*): client credentials — works out of box
- User actions (fave, collections CRUD, gallery CRUD, messages): authorization-code flow — requires OAuth login

If user actions fail with auth error, Fahad needs to authenticate. Check MCP server logs for auth URL.

## Common workflows

- **Find art:** `browse_popular` or `browse_tags` with a tag name
- **Check an artist:** `gallery_all` with username, or `browse_user_journals`
- **Save favorites:** `collections_fave` with deviation ID (requires auth)
- **Review notifications:** `messages_feedback` + `messages_mentions` (requires auth)
