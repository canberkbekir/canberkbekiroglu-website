namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphImageAlt
{
	public string OpenGraph { get; set; } = string.Empty;
	public string Twitter { get; set; } = string.Empty;

	public OpenGraphImageAlt()
	{
	}

	public OpenGraphImageAlt(string value)
	{
		OpenGraph = value;
		Twitter = value;
	}

	public static implicit operator OpenGraphImageAlt(string value)
	{
		return new OpenGraphImageAlt(value);
	}
}
