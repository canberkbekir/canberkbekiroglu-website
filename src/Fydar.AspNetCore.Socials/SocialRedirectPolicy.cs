using Fydar.AspNetCore.OpenGraph;

namespace Fydar.AspNetCore.Socials;

public class SocialRedirectPolicy
{
	public required string Destination { get; set; } = string.Empty;
	public OpenGraphModel? Model { get; set; }
}
