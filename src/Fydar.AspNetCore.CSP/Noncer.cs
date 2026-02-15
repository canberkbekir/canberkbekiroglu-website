using System;

namespace Fydar.AspNetCore.CSP;

public class Noncer
{
	private string? nonce;

	public string Nonce
	{
		get
		{
			if (string.IsNullOrEmpty(nonce))
			{
				using var rng = System.Security.Cryptography.RandomNumberGenerator.Create();
				byte[] nonceBytes = new byte[32];
				rng.GetBytes(nonceBytes);
				nonce = Convert.ToBase64String(nonceBytes);
			}
			return nonce;
		}
	}
}
