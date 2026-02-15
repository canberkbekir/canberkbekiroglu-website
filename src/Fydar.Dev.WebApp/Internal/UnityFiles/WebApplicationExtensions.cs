using Microsoft.AspNetCore.StaticFiles;
using Microsoft.Extensions.FileProviders;

namespace Fydar.Dev.WebApp.Internal.UnityFiles;

internal static class WebApplicationExtensions
{
	public static void UseStaticUnityFiles(this WebApplication app)
	{
		var provider = new FileExtensionContentTypeProvider();
		provider.Mappings.Remove(".br");
		provider.Mappings.Clear();
		provider.Mappings[".js"] = "application/javascript";
		provider.Mappings[".br"] = "application/octet-stream";
		provider.Mappings[".data"] = "application/octet-stream";
		provider.Mappings[".bank"] = "application/octet-stream";

		app.UseStaticFiles(new StaticFileOptions
		{
			FileProvider = new PhysicalFileProvider(
				Path.Combine(app.Environment.ContentRootPath, "precompressed")),
			ContentTypeProvider = provider,
			OnPrepareResponse = context =>
			{
				var headers = context.Context.Response.Headers;

				if (context.File.Name.EndsWith(".br", StringComparison.OrdinalIgnoreCase))
				{
					headers.ContentEncoding = "br";

					if (context.File.Name.Contains(".wasm", StringComparison.OrdinalIgnoreCase))
					{
						headers.ContentType = "application/wasm";
					}
					else if (context.File.Name.Contains(".js", StringComparison.OrdinalIgnoreCase))
					{
						headers.ContentType = "application/javascript";
					}
					else if (context.File.Name.Contains(".data", StringComparison.OrdinalIgnoreCase))
					{
						headers.ContentType = "application/octet-stream";
					}
				}

				if (context.File.Name.EndsWith(".data", StringComparison.OrdinalIgnoreCase))
				{
					headers.ContentType = "application/octet-stream";
				}
			}
		});
	}
}
