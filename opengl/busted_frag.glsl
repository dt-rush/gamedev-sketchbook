#version 330

#define BYPASS 1
#define PIXEL_SIZE 3.0
#define constAtt 1.0
#define linAtt 0.1
#define quadAtt 0.1

uniform float aliveness;
uniform int numPointLights;
uniform samplerBuffer pointLights;

out vec4 outColor;



void main()
{
vec3 fragmentColor = vec3(1);
	float ambient = 0.1;
	// Initialize the lighting or color value to zero
    vec3 totalLight = vec3(ambient);
    for (int i = 0; i < numPointLights; i++) {
        vec3 lightPos = texelFetch(pointLights, 2*i).xyz;
        vec3 lightColor = texelFetch(pointLights, 2*i+1).xyz;
        vec3 lightDir = normalize(lightPos - gl_FragCoord);
        float dist = length(lightPos - gl_FragCoord);
        float attenuation = 1.0 / (constAtt + linAtt * dist + quadAtt * dist * dist);
        vec3 diffuse = max(dot(normal, lightDir), 0.0) * lightColor;
        // vec3 specular = pow(max(dot(viewDir, reflect(-lightDir, normal)), 0.0), shininess) * lightColor * specularStrength;
        vec3 specular = vec3(0);
        vec3 light = (diffuse + specular) * attenuation;
        totalLight += light;
    }

	// compute greyscale brightness
	float brightness = (totalLight.r + totalLight.g + totalLight.b) / 3;

	if (BYPASS == 0) {

		vec2 xy = gl_FragCoord.xy;
		vec2 pixel = mod(xy/PIXEL_SIZE, 4.0);

		int x = int(pixel.x);
		int y = int(pixel.y);

		bool result = false;
		if (x == 0 && y == 0) result = brightness > 16.0/17.0;
		else if (x == 2 && y == 2) result = brightness > 15.0/17.0;
		else if (x == 2 && y == 0) result = brightness > 14.0/17.0;
		else if (x == 0 && y == 2) result = brightness > 13.0/17.0;
		else if (x == 1 && y == 1) result = brightness > 12.0/17.0;
		else if (x == 3 && y == 3) result = brightness > 11.0/17.0;
		else if (x == 3 && y == 1) result = brightness > 10.0/17.0;
		else if (x == 1 && y == 3) result = brightness > 09.0/17.0;
		else if (x == 1 && y == 0) result = brightness > 08.0/17.0;
		else if (x == 3 && y == 2) result = brightness > 07.0/17.0;
		else if (x == 3 && y == 0) result = brightness > 06.0/17.0;
		else if (x == 0 && y == 1) result =	brightness > 05.0/17.0;
		else if (x == 1 && y == 2) result = brightness > 04.0/17.0;
		else if (x == 2 && y == 3) result = brightness > 03.0/17.0;
		else if (x == 2 && y == 1) result = brightness > 02.0/17.0;
		else if (x == 0 && y == 3) result = brightness > 01.0/17.0;

		vec3 onOff = vec3(result);
		outColor = vec4(mix(onOff, fragmentColor, clamp(aliveness/5, 0, 1)), 1.0);
	} else if (BYPASS == 1) {
		outColor = vec4(vec3(brightness), 1.0);
	} else if (BYPASS == 2) {
		outColor = vec4(1.0);
	}

}
