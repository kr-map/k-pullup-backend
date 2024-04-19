import logout from "@/api/auth/logout";
import { useMutation, useQueryClient } from "@tanstack/react-query";

const useLogout = () => {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: logout,
    onError: (error) => {
      console.log(error);
    },
    onSuccess: () => {
      window.location.reload();
    },
  });
};

export default useLogout;
