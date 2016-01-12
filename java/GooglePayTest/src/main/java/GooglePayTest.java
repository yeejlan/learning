
import com.google.Security;

public class GooglePayTest{ 

	public static void main(String[] args) {  
		boolean verify = Security.verifyPurchase("MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCoNVF/DoNOpY7vsWe4Nt+OUM4pZaBcG7ugTg0T347455ipLCH60YH1pF4N1AKDfGuG5HtGXCzOEH6KcLY0JKA5kEnzVwPXrovg1s9oAUp+f7+SDnymf2KbchraTHvtP718N3oxSaVl0JcxksfFIkQvS3nnQI6YoDPHpn9l1sBBlQIDAQAB",
		 "my test data",
		 "V18LTKZ3NjZyqkpYeqXByHXGHoZI+GpXSrZBEY43XtVcDdQype5x1RrIEbarjXi3jwBbcWStnyGS3858Bz5snLJI8SCa/cAROTKomWq2fqLMcMfedQga3uaS4BSYUtGOl14Rw3x0q9Z+DdNSKzIcq3mrIFON126psfLBaXUU6GM=");

		System.out.println(verify); 
	}

}
