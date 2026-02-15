namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphCanonicalUrl
{
	public string Html { get; set; } = string.Empty;
	public string OpenGraph { get; set; } = string.Empty;
	public string Twitter { get; set; } = string.Empty;

	public OpenGraphCanonicalUrl()
	{
	}

	public OpenGraphCanonicalUrl(string value)
	{
		Html = value;
		OpenGraph = value;
		Twitter = value;
	}

	public static implicit operator OpenGraphCanonicalUrl(string value)
	{
		return new OpenGraphCanonicalUrl(value);
	}
}
