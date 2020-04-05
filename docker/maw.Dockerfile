FROM mcr.microsoft.com/dotnet/core/sdk:3.0 AS build-env
#install node and npm
RUN curl --silent --location https://deb.nodesource.com/setup_10.x | bash -
RUN apt-get install --yes nodejs

WORKDIR /app

# Copy csproj and restore as distinct layers
COPY web/meetingadmin/*.csproj ./
RUN dotnet restore

# Copy everything else and build
COPY web/meetingadmin ./
RUN dotnet publish -c Release -o out

# Build runtime image
FROM mcr.microsoft.com/dotnet/core/aspnet:3.0
WORKDIR /app
COPY --from=build-env /app/out .
ENTRYPOINT ["dotnet", "meetingadmin.dll"]

