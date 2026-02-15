namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphImageUrl
{
	public string OpenGraphUrl { get; set; } = string.Empty;
	public string OpenGraphSecureUrl { get; set; } = string.Empty;
	public string TwitterUrl { get; set; } = string.Empty;

	public OpenGraphImageUrl()
	{
	}

	public OpenGraphImageUrl(string value)
	{
		OpenGraphUrl = value;
		OpenGraphSecureUrl = value;
		TwitterUrl = value;
	}

	public static implicit operator OpenGraphImageUrl(string value)
	{
		return new OpenGraphImageUrl(value);
	}
}
