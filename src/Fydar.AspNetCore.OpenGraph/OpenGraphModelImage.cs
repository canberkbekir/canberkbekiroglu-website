namespace Fydar.AspNetCore.OpenGraph;

public class OpenGraphModelImage : IOpenGraphObject
{
	public OpenGraphImageUrl Url { get; set; } = new();
	public OpenGraphImageAlt Alt { get; set; } = new();
}
