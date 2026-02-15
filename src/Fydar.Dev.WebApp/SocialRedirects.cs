using Fydar.AspNetCore.Socials;

namespace Fydar.Dev.WebApp;

internal static class SocialRedirects
{
	public static void MapSocialRedirects(
		this IEndpointRouteBuilder app)
	{
		// GitHub
		app.MapSocialRedirect("/github", new()
		{
			Destination = "https://github.com/Fydar",
			Model = new()
			{
				Title = "Fydar on GitHub",
				Description = "Explore my open-source projects, code repositories, and contributions to the .NET ecosystem."
			}
		});

		// LinkedIn
		app.MapSocialRedirect("/linkedin", new()
		{
			Destination = "https://www.linkedin.com/in/fydar/",
			Model = new()
			{
				Title = "Connect with Fydar on LinkedIn",
				Description = "Professional profile, project history, and technical insights.",
			}
		});

		// YouTube
		app.MapSocialRedirect("/youtube", new()
		{
			Destination = "https://youtube.com/@fydar",
			Model = new()
			{
				Title = "Fydar on YouTube",
				Description = "Technical tutorials, project demos, and deep dives into modern software engineering."
			}
		});

		// Discord
		app.MapSocialRedirect("/discord", new()
		{
			Destination = "https://discord.com/users/172972361954492416",
			Model = new()
			{
				Title = "Connect with Fydar on Discord",
				Description = "Let's connect."
			}
		});

		// Gravatar
		app.MapSocialRedirect("/gravatar", new()
		{
			Destination = "https://gravatar.com/fydardev",
			Model = new()
			{
				Title = "Fydar's Global Profile",
				Description = "The centralized identity and avatar used across the web."
			}
		});

		// Reddit
		app.MapSocialRedirect("/reddit", new()
		{
			Destination = "https://www.reddit.com/user/Fydarus/",
			Model = new()
			{
				Title = "Fydar on Reddit",
				Description = "Engaging in communities focused on software engineering and indie development."
			}
		});

		// Instagram
		app.MapSocialRedirect("/instagram", new()
		{
			Destination = "https://www.instagram.com/fydar.dev/",
			Model = new()
			{
				Title = "Fydar on Instagram",
				Description = "Personal Instagram."
			}
		});

		// Threads
		app.MapSocialRedirect("/threads", new()
		{
			Destination = "https://www.threads.com/@fydar.dev",
			Model = new()
			{
				Title = "Fydar on Threads",
				Description = "Text-based updates and community engagement from the Meta ecosystem."
			}
		});

		// BlueSky
		app.MapSocialRedirect("/bluesky", new()
		{
			Destination = "https://bsky.app/profile/fydar.bsky.social",
			Model = new()
			{
				Title = "Fydar on BlueSky",
				Description = "Decentralized social media updates on tech, code, and life."
			}
		});

		// Twitter
		app.MapSocialRedirect("/twitter", new()
		{
			Destination = "https://twitter.com/Fydarus",
			Model = new()
			{
				Title = "Fydar on Twitter",
				Description = "Technical insights, .NET updates, and software development thoughts in 280 characters or less."
			}
		});

		// Spotify
		app.MapSocialRedirect("/spotify", new()
		{
			Destination = "https://open.spotify.com/user/fydar",
			Model = new()
			{
				Title = "Fydar's Playlists",
				Description = "Curation of music for focused deep-work and development sessions."
			}
		});

		// Steam
		app.MapSocialRedirect("/steam", new()
		{
			Destination = "https://steamcommunity.com/id/Fydar/",
			Model = new()
			{
				Title = "Fydar on Steam",
				Description = "Reviewing games, tracking achievements, and connecting to play."
			}
		});

		// CurseForge
		app.MapSocialRedirect("/curseforge", new()
		{
			Destination = "https://www.curseforge.com/members/fydar",
			Model = new()
			{
				Title = "Fydar's Mods on CurseForge",
				Description = "Downloadable extensions and modifications for popular gaming titles."
			}
		});

		// Modrinth
		app.MapSocialRedirect("/modrinth", new()
		{
			Destination = "https://modrinth.com/user/fydar",
			Model = new()
			{
				Title = "Fydar on Modrinth",
				Description = "A collection of open-source and high-performance game modifications."
			}
		});

		// TikTok
		app.MapSocialRedirect("/tiktok", new()
		{
			Destination = "https://tiktok.com/@fydar",
			Model = new()
			{
				Title = "Fydar on TikTok",
				Description = "Short-form dev logs and rapid-fire technical tips."
			}
		});

		// Docker
		app.MapSocialRedirect("/docker", new()
		{
			Destination = "https://hub.docker.com/u/fydar",
			Model = new()
			{
				Title = "Fydar's Docker Hub",
				Description = "Containerized applications and optimized images for .NET deployments."
			}
		});

		// Notion
		app.MapSocialRedirect("/notion", new()
		{
			Destination = "https://notion.so/fydar",
			Model = new()
			{
				Title = "Fydar's Knowledge Base",
				Description = "Public documentation, roadmaps, and project management workspaces."
			}
		});

		// Twitch
		app.MapSocialRedirect("/twitch", new()
		{
			Destination = "https://twitch.tv/fydar",
			Model = new()
			{
				Title = "Fydar Live on Twitch",
				Description = "Join my live coding sessions to see real-time software builds and C# development."
			}
		});

		// Itch.IO
		app.MapSocialRedirect("/itch", new()
		{
			Destination = "https://fydar.itch.io",
			Model = new()
			{
				Title = "Fydar's Games & Tools",
				Description = "Browse my indie game projects, experimental software, and digital tools on Itch.io."
			}
		});

		// Unity Asset Store
		app.MapSocialRedirect("/unity", new()
		{
			Destination = "https://assetstore.unity.com/publishers/13236",
			Model = new()
			{
				Title = "Fydar's Unity Assets",
				Description = "High-quality scripts and tools designed to accelerate your game development workflow."
			}
		});

		// SketchFab
		app.MapSocialRedirect("/sketchfab", new()
		{
			Destination = "https://sketchfab.com/fydar",
			Model = new()
			{
				Title = "Fydar's 3D Gallery",
				Description = "Interactive 3D models, assets, and environmental art for game development."
			}
		});

		// ArtStation
		app.MapSocialRedirect("/artstation", new()
		{
			Destination = "https://www.artstation.com/fydar",
			Model = new()
			{
				Title = "Fydar's Portfolio on ArtStation",
				Description = "Visual design, digital art, and creative direction projects."
			}
		});

		// OpenCollective
		app.MapSocialRedirect("/opencollective", new()
		{
			Destination = "https://opencollective.com/fydar",
			Model = new()
			{
				Title = "Fydar on OpenCollective",
				Description = "Helping to sustain open-source development and community-driven software projects."
			}
		});

		// Stack Overflow
		app.MapSocialRedirect("/stackoverflow", new()
		{
			Destination = "https://stackoverflow.com/users/9726948/fydar",
			Model = new()
			{
				Title = "Fydar's StackOverflow Profile",
				Description = "Review my technical contributions, answers, and reputation within the StackOverflow community."
			}
		});

		// GitLab
		app.MapSocialRedirect("/gitlab", new()
		{
			Destination = "https://gitlab.com/fydar",
			Model = new()
			{
				Title = "Fydar on GitLab",
				Description = "DevOps, CI/CD pipelines, and private repository management."
			}
		});

		// Trello
		app.MapSocialRedirect("/trello", new()
		{
			Destination = "https://trello.com/fydar",
			Model = new()
			{
				Title = "Fydar's Project Boards on Trello",
				Description = "Transparent task tracking and development workflow management."
			}
		});

		// XBOX
		app.MapSocialRedirect("/xbox", new()
		{
			Destination = "https://www.xbox.com/en-GB/play/user/Fydar6121",
			Model = new()
			{
				Title = "Fydar's Xbox Profile",
				Description = "Gaming activity and achievements across the Microsoft ecosystem."
			}
		});

		// Zoom
		app.MapSocialRedirect("/zoom", new()
		{
			Destination = "https://community.zoom.com/t5/user/viewprofilepage/user-id/771284",
			Model = new()
			{
				Title = "Fydar's Zoom Room",
				Description = "Personal meeting space for technical consultations and remote collaboration."
			}
		});

		// Atlassian / Community
		// app.MapSocialRedirect("/atlassian", new()
		// {
		// 	Destination = "https://community.atlassian.com/t5/user/viewprofilepage/user-id/your-id",
		// 	Model = new()
		// 	{
		// 		Title = "Fydar on Atlassian Community",
		// 		Description = "Insights and discussions on Jira, Confluence, and team collaboration tools."
		// 	}
		// });
		//
		// // Meta Stack Overflow
		// app.MapSocialRedirect("/metastackoverflow", new()
		// {
		// 	Destination = "https://meta.stackoverflow.com/users/9726948/fydar",
		// 	Model = new()
		// 	{
		// 		Title = "Fydar on Meta Stack Overflow",
		// 		Description = "Participating in discussions regarding the future of the developer community."
		// 	}
		// });
		// 
		// // Stack Exchange
		// app.MapSocialRedirect("/stackexchange", new()
		// {
		// 	Destination = "https://stackexchange.com/users/13481916/fydar",
		// 	Model = new()
		// 	{
		// 		Title = "Fydar's StackExchange Profile",
		// 		Description = "Review my technical contributions, answers, and reputation within the StackExchange community."
		// 	}
		// });
		// 
		// // Game Dev Stack Exchange
		// app.MapSocialRedirect("/gamedevstackexchange", new()
		// {
		// 	Destination = "https://gamedev.stackexchange.com/users/142990/fydar",
		// 	Model = new()
		// 	{
		// 		Title = "Fydar's Game Development StackExchange Profile",
		// 		Description = "Review my technical contributions, answers, and reputation within the StackExchange community."
		// 	}
		// });
		// 
		// // Unity Discussions
		// app.MapSocialRedirect("/unitydiscussions", new()
		// {
		// 	Destination = "https://discussions.unity.com/u/fydar",
		// 	Model = new()
		// 	{
		// 		Title = "Fydar on the Unity Community",
		// 		Description = "Active member of the Unity engine discussion forums and technical troubleshooting." }
		// });
	}
}
