namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphModel
{
	public OpenGraphTitle Title { get; set; } = new();
	public OpenGraphDescription Description { get; set; } = new();
	public OpenGraphCanonicalUrl CanonicalUrl { get; set; } = new();
	public IOpenGraphObject[] Properties { get; set; } = [];
}
