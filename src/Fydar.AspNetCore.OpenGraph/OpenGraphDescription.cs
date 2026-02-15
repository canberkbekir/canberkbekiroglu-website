namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphDescription
{
	public string Default { get; set; } = string.Empty;
	public string OpenGraph { get; set; } = string.Empty;
	public string Twitter { get; set; } = string.Empty;

	public OpenGraphDescription()
	{
	}

	public OpenGraphDescription(string value)
	{
		Default = value;
		OpenGraph = value;
		Twitter = value;
	}

	public static implicit operator OpenGraphDescription(string value)
	{
		return new OpenGraphDescription(value);
	}
}
