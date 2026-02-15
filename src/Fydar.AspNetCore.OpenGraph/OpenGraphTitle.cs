namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphTitle
{
	public string Html { get; set; } = string.Empty;
	public string OpenGraph { get; set; } = string.Empty;
	public string Twitter { get; set; } = string.Empty;

	public OpenGraphTitle()
	{
	}

	public OpenGraphTitle(string value)
	{
		Html = value;
		OpenGraph = value;
		Twitter = value;
	}

	public static implicit operator OpenGraphTitle(string value)
	{
		return new OpenGraphTitle(value);
	}
}
