// The placeholder image URL to use with users whose profile picture is not set.
const placeholder = "/public/placeholder.jpg";

// The go-instaman endpoint that serves pictures.
const proxyURL = "/instaman/instagram/picture";

type Props = {
  onClick?: React.MouseEventHandler;
  height?: string;
  pictureURL?: URL;
  width?: string;
};

// Return the "proxified" URL for the given picture URL.
const makeProxiedURL = (pictureURL: URL): string => {
  const proxied = new URL(proxyURL, window.location.origin);
  proxied.searchParams.append("pictureURL", pictureURL.toString());

  return proxied.toString();
};

// Render an <img> with the given user's picture. The image's URL will be proxified through go-instaman to avoid CORS blocking.
const UserProfilePicture = ({ onClick, height, pictureURL, width }: Props) => {
  return (
    <img
      height={height}
      onClick={onClick}
      src={pictureURL ? makeProxiedURL(pictureURL) : placeholder}
      style={{
        // Use cursor:pointer whenever an onClick handler is specified.
        cursor: onClick ? "pointer" : undefined,
      }}
      width={width}
    />
  );
};

export default UserProfilePicture;
